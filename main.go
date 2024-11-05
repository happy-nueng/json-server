package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/gofiber/fiber/v2"
	"gopkg.in/yaml.v2"
)

type RouteConfig struct {
	Method       string `yaml:"method"`
	Route        string `yaml:"route"`
	ResponseFile string `yaml:"response_file"`
}

type Config struct {
	Routes     []RouteConfig `yaml:"routes"`
	ServerPort int           `yaml:"server_port"`
}

func loadResponseFromFile(filePath string) (interface{}, error) {
	fileData, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("ไม่สามารถอ่านไฟล์ %s ได้: %v", filePath, err)
	}

	var response interface{}
	if err := json.Unmarshal(fileData, &response); err != nil {
		return nil, fmt.Errorf("ไม่สามารถแปลง JSON ได้: %v", err)
	}
	return response, nil
}

func filterResponseData(responseData interface{}, queryParams map[string]string) []interface{} {
	var response []interface{}
	switch data := responseData.(type) {
	case map[string]interface{}:
		for key, value := range data {
			if qValue, ok := queryParams[key]; ok && fmt.Sprint(value) == qValue {
				response = append(response, value)
			}
		}
	case []interface{}:
		for _, item := range data {
			if itemMap, ok := item.(map[string]interface{}); ok {
				for key, value := range itemMap {
					if qValue, ok := queryParams[key]; ok && fmt.Sprint(value) == qValue {
						response = append(response, itemMap)
					}
				}
			}
		}
	}
	return response
}

func main() {
	app := fiber.New()

	configData, err := os.ReadFile("config.yaml")
	if err != nil {
		log.Fatalf("ไม่สามารถอ่านไฟล์ config.yaml ได้: %v", err)
	}

	var config Config
	if err := yaml.Unmarshal(configData, &config); err != nil {
		log.Fatalf("ไม่สามารถแปลง YAML เป็นโครงสร้าง Go ได้: %v", err)
	}

	for _, route := range config.Routes {
		responseData, err := loadResponseFromFile(route.ResponseFile)
		if err != nil {
			log.Fatalf("Error loading response file for route %s: %v", route.Route, err)
		}

		handler := func(c *fiber.Ctx) error {
			queryParams := c.Queries()
			var response []interface{}
			if len(queryParams) > 0 {
				response = filterResponseData(responseData, queryParams)
			} else {
				response = responseData.([]interface{})
			}
			return c.JSON(response)
		}

		switch route.Method {
		case "GET":
			app.Get(route.Route, handler)
		case "POST":
			app.Post(route.Route, func(c *fiber.Ctx) error {
				return c.JSON(responseData)
			})
		default:
			log.Printf("Method %s ไม่รองรับในขณะนี้", route.Method)
		}
	}

	for _, route := range app.GetRoutes() {
		fmt.Printf("Route: %s %s\n", route.Method, route.Path)
	}

	fmt.Printf("Server เริ่มทำงานที่ http://localhost:%v\n", config.ServerPort)
	app.Listen(fmt.Sprintf(":%v", config.ServerPort))
}
