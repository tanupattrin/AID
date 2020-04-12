// Copyright (c) 2020 Xiaozhe Yao & AICAMP.CO.,LTD
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package daemon

import (
	"net/http"
	"path/filepath"
	"strings"

	"github.com/autoai-org/aid/components/cmd/pkg/entities"
	"github.com/autoai-org/aid/components/cmd/pkg/runtime"
	"github.com/autoai-org/aid/components/cmd/pkg/utilities"
	"github.com/gin-gonic/gin"
)

// getPackages returns all packages
// GET /packages -> returns packages
func getPackages(c *gin.Context) {
	packages := entities.FetchPackages()
	c.JSON(http.StatusOK, packages)
}

// getImages returns all images that has been built
// GET /images -> returns all containers
func getImages(c *gin.Context) {
	images := entities.FetchImages()
	c.JSON(http.StatusOK, images)
}

// getSolvers returns all solvers
// GET /solvers -> returns all solvers
func getSolvers(c *gin.Context) {
	solvers := entities.FetchSolvers()
	c.JSON(http.StatusOK, solvers)
}

// getContainers returns all containers
// GET /containers -> returns all containers
func getContainers(c *gin.Context) {
	containers := entities.FetchContainers()
	c.JSON(http.StatusOK, containers)
}

// getSolverDockerfile returns current dockerfile of the solver
// GET /packages/:vendorName/:packageName/:solverName/dockerfile
func getDockerfileContent(c *gin.Context) {
	packageFolder := filepath.Join(utilities.GetBasePath(), "models", c.Param("vendorName"), c.Param("packageName"))
	dockerFilename := "docker_" + c.Param("solverName")
	dockerFilePath := filepath.Join(packageFolder, dockerFilename)
	fileContent, err := utilities.ReadFileContent(dockerFilePath)
	if err != nil {
		c.JSON(http.StatusOK, messageResponse{
			Code: 404,
			Msg:  err.Error(),
		})
	} else {
		c.JSON(http.StatusOK, messageResponse{
			Code: 200,
			Msg:  fileContent,
		})
	}
}

// modifySolverDockerfile modifies the content of the dockerfile
// POST /packages/:vendorName/:packageName/:solverName/dockerfile
func modifySolverDockerfile(c *gin.Context) {
	var request modifySolverDockerfileRequest
	c.BindJSON(&request)
	packageFolder := filepath.Join(utilities.GetBasePath(), "models", c.Param("vendorName"), c.Param("packageName"))
	dockerFilename := "docker_" + c.Param("solverName")
	dockerFilePath := filepath.Join(packageFolder, dockerFilename)
	err := utilities.WriteContentToFile(dockerFilePath, request.Content)
	if err != nil {
		c.JSON(http.StatusOK, messageResponse{
			Code: 404,
			Msg:  err.Error(),
		})
	} else {
		c.JSON(http.StatusOK, messageResponse{
			Code: 200,
			Msg:  "Successfully Modified",
		})
	}
}

// getMetaInfo returns all meta information about the package
// GET /:vendorName/:packageName/meta -> {readme, aid.toml, pretrained.toml}
func getMetaInfo(c *gin.Context) {
	var solvers entities.Solvers
	var pretraineds entities.Pretraineds
	var readmeContent string
	var requirements string
	packageFolder := filepath.Join(utilities.GetBasePath(), "models", c.Param("vendorName"), c.Param("packageName"))
	aidtomlString, err := utilities.ReadFileContent(filepath.Join(packageFolder, "aid.toml"))
	if err == nil {
		solvers = entities.LoadSolversFromConfig(aidtomlString)
	}
	pretrainedtomlString, err := utilities.ReadFileContent(filepath.Join(packageFolder, "pretrained.toml"))
	if err == nil {
		pretraineds = entities.LoadPretrainedsFromConfig(pretrainedtomlString)
	}
	readmeContent, err = utilities.ReadFileContent(filepath.Join(packageFolder, "README.md"))
	requirements, err = utilities.ReadFileContent(filepath.Join(packageFolder, "requirements.txt"))
	c.JSON(http.StatusOK, metaResponse{
		Solvers: solvers, Pretraineds: pretraineds, Readme: readmeContent, Requirements: requirements,
	})
}

// installPackage performs installation
// PUT /packages
func installPackage(c *gin.Context) {
	var request installPackageRequest
	c.BindJSON(&request)
	targetPath := filepath.Join(utilities.GetBasePath(), "models")
	err := runtime.InstallPackage(request.RemoteURL, targetPath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, messageResponse{
			Msg: err.Error(),
		})
	}
	c.JSON(http.StatusOK, messageResponse{
		Code: 200,
		Msg:  "submitted success",
	})
}

// createContainers will create a container by using the given image
// PUT /images/:imageId/containers
func createSolverContainer(c *gin.Context) {
	imageID := c.Param("imageId")
	dockerClient := runtime.NewDockerRuntime()
	_, err := dockerClient.Create(imageID, "8081")
	if err != nil {
		c.JSON(http.StatusInternalServerError, messageResponse{
			Msg: err.Error(),
		})
	} else {
		c.JSON(http.StatusOK, messageResponse{
			Code: 200,
			Msg:  "submitted success",
		})
	}
}

// startContainers will start the container as daemon
// PUT /containers/:containerId/run
func startSolverContainer(c *gin.Context) {
	containerID := c.Param("containerId")
	dockerClient := runtime.NewDockerRuntime()
	err := dockerClient.Start(containerID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, messageResponse{
			Msg: err.Error(),
		})
	} else {
		c.JSON(http.StatusOK, messageResponse{
			Code: 200,
			Msg:  "submitted success",
		})
	}
}

// buildPackages build a new image
// PUT /:vendorName/:packageName/:solverName/images
func buildSolverImage(c *gin.Context) {
	dockerClient := runtime.NewDockerRuntime()
	packageFolder := filepath.Join(utilities.GetBasePath(), "models", c.Param("vendorName"), c.Param("packageName"))
	tomlString, err := utilities.ReadFileContent(filepath.Join(packageFolder, "aid.toml"))
	if err != nil {
		utilities.CheckError(err, "Cannot open file "+filepath.Join(packageFolder, "aid.toml"))
	}
	solvers := entities.LoadSolversFromConfig(tomlString)
	runtime.RenderRunnerTpl(packageFolder, solvers)
	packageInfo := entities.LoadPackageFromConfig(tomlString)
	solverName := c.Param("solverName")
	var imageName string
	var log entities.Log
	imageName = packageInfo.Package.Vendor + "-" + packageInfo.Package.Name + "-" + solverName
	// Check if docker file exists
	if !utilities.IsExists(filepath.Join(packageFolder, "docker_"+solverName)) {
		runtime.RenderDockerfile(solverName, packageFolder)
	}
	log, err = dockerClient.Build(strings.ToLower(imageName), filepath.Join(packageFolder, "docker_"+solverName))
	if err != nil {
		c.JSON(http.StatusInternalServerError, messageResponse{
			Code: 500,
			Msg:  err.Error(),
		})
	} else {
		c.JSON(http.StatusOK, gin.H{
			"code":  200,
			"logid": log.ID,
		})
	}
}

// getEnvironmentVariables will return requested environment variables
// GET /:vendorName/:packageName/envs?env={all, dev, test, prod}
func getEnvironmentVariables(c *gin.Context) {
	var req queryEnvironmentVariablesRequest
	c.Bind(&req)
	reqPackage := entities.GetPackage(c.Param("vendorName"), c.Param("packageName"))
	envs := GetEnvironmentVariablesbyPackageID(reqPackage.ID, req.Env)
	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"results": envs,
	})
}

// setEnvironmentVariables will set a new environment variable
// PUT /:vendorName/:packageName/envs/:envName
func setEnvironmentVariables(c *gin.Context) {
	reqPackage := entities.GetPackage(c.Param("vendorName"), c.Param("packageName"))
	var req newEnvironmentVariableRequest
	c.BindJSON(&req)
	env := entities.EnvironmentVariable{Environment: c.Param("envName"), PackageID: reqPackage.ID, Key: req.Key, Value: req.Value}
	err := env.Save()
	if err != nil {
		c.JSON(http.StatusInternalServerError, messageResponse{
			Code: 500,
			Msg:  err.Error(),
		})
	} else {
		c.JSON(http.StatusOK, gin.H{
			"code":  200,
			"logid": log.ID,
		})
	}
}
