package main

import (
	"context"
	"fmt"
	"net/http"
	"strconv"

	"github.com/byuoitav/atlona-driver/atlona"
	"github.com/byuoitav/atlona-switcher-microservice/handlers5x1"
	"github.com/byuoitav/common"
	"github.com/byuoitav/common/status"
	"github.com/labstack/echo"
)

func main() {
	port := ":8080"
	router := common.NewRouter()

	Atlonavs, _ := atlona.CreateVideoSwitcher(context.Background(), "10.66.9.2", "Steve", "steve")

	// Functionality Endpoints
	write := router.Group("")
	// 5x1 functionality Endpoints
	write.GET("/:address/output/:output/input/:input", func(ectx echo.Context) error {
		fmt.Println("Attempting to change input")
		input := ectx.Param("input")
		output := ectx.Param("output")
		intInput, nerr := strconv.Atoi(input)
		if nerr != nil {
			return ectx.String(http.StatusInternalServerError, nerr.Error())
		}
		if intInput > 5 {
			return ectx.String(http.StatusInternalServerError, "Invalid Input")
		}
		ctx := ectx.Request().Context()
		er := Atlonavs.SetInputByOutput(ctx, output, input)
		if er != nil {
			return ectx.String(http.StatusInternalServerError, er.Error())
		}

		return ectx.JSON(http.StatusOK, status.Input{
			Input: fmt.Sprintf("%v:1", input),
		})
	})
	write.GET("/:address/volume/:level/:input", handlers5x1.SetVolume)
	write.GET("/:address/output/:output/input", func(ectx echo.Context) error {
		output := ectx.Param("output")
		ctx := ectx.Request().Context()
		input, err := Atlonavs.GetInputByOutput(ctx, output)
		if err != nil {
			return ectx.String(http.StatusInternalServerError, err.Error())
		}

		return ectx.JSON(http.StatusOK, status.Input{
			Input: fmt.Sprintf("%s:1", input),
		})
	})
	write.GET("/:address/volume/:input", handlers5x1.GetVolume)
	write.GET("/:address/muteStatus/:input", handlers5x1.GetMute)

	server := http.Server{
		Addr:           port,
		MaxHeaderBytes: 1024 * 10,
	}

	router.StartServer(&server)
}
