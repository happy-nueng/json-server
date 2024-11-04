package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/gofiber/fiber/v2"
	"gopkg.in/yaml.v2"
)

// โครงสร้างข้อมูลสำหรับ route config
type RouteConfig struct {
	Method       string `yaml:"method"`
	Route        string `yaml:"route"`
	ResponseFile string `yaml:"response_file"`
}

type Config struct {
	Routes     []RouteConfig `yaml:"routes"`
	ServerPort int           `yaml:"server_port"`
}

func loadResponseFromFile(filename string) (interface{}, error) {
	// อ่านไฟล์ JSON ตามชื่อไฟล์ที่กำหนด
	fileData, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("ไม่สามารถอ่านไฟล์ %s ได้: %v", filename, err)
	}

	// แปลงข้อมูล JSON ให้เป็นรูปแบบโครงสร้าง Go (interface{})
	var response interface{}
	if err := json.Unmarshal(fileData, &response); err != nil {
		return nil, fmt.Errorf("ไม่สามารถแปลง JSON ได้: %v", err)
	}
	return response, nil
}

func main() {
	app := fiber.New()

	// อ่านไฟล์ config.yaml
	configData, err := os.ReadFile("config.yaml")
	if err != nil {
		log.Fatalf("ไม่สามารถอ่านไฟล์ config.yaml ได้: %v", err)
	}

	// แปลงข้อมูลจาก YAML เป็นโครงสร้าง Go
	var config Config
	if err := yaml.Unmarshal(configData, &config); err != nil {
		log.Fatalf("ไม่สามารถแปลง YAML เป็นโครงสร้าง Go ได้: %v", err)
	}

	// สร้าง route ตามที่กำหนดใน config.yaml
	for _, route := range config.Routes {
		// โหลด response file
		responseData, err := loadResponseFromFile(route.ResponseFile)
		if err != nil {
			log.Fatalf("Error loading response file for route %s: %v", route.Route, err)
		}

		// ตรวจสอบ Method และสร้าง route ที่รองรับพารามิเตอร์
		switch route.Method {
		case "GET":
			app.Get(route.Route, func(c *fiber.Ctx) error {
				// ดึง query parameters จาก GET เช่น `?name=John&id=1`
				parameters := c.Queries()
				var response []interface{}
				if len(parameters) > 0 {
					switch responseData := responseData.(type) {
					case map[string]interface{}:
						// ตรวจสอบว่า responseData เป็น map[string]interface{}
						for key, value := range responseData {
							for qKey, qValue := range parameters {
								if key == qKey && fmt.Sprint(value) == qValue {
									response = append(response, value)
								}
							}
						}
					case []interface{}:
						// ตรวจสอบว่า responseData เป็น []interface{}
						for i, values := range responseData {
							switch values := values.(type) {
							case map[string]interface{}:
								// ตรวจสอบว่า value เป็น map[string]interface{}
								for key, value := range values {
									for qKey, qValue := range parameters {
										if key == qKey && fmt.Sprint(value) == qValue {
											response = append(response, values)
										}
									}
								}
							case string:
								// ตรวจสอบว่า value เป็น string
								fmt.Printf("Parameter: %d = %v\n", i, values)
							}
						}
					}
				} else {
					response = responseData.([]interface{})
				}

				// ตอบกลับข้อมูล JSON
				return c.JSON(response)
			})
		case "POST":
			app.Post(route.Route, func(c *fiber.Ctx) error {
				// ตอบกลับข้อมูล JSON
				return c.JSON(responseData)
			})
		default:
			log.Printf("Method %s ไม่รองรับในขณะนี้", route.Method)
		}
	}

	routes := app.GetRoutes()
	for _, route := range routes {
		fmt.Printf("Route: %s %s\n", route.Method, route.Path)
	}

	// port ที่ server เริ่มทำงาน
	fmt.Printf("Server เริ่มทำงานที่ http://localhost:%v\n", config.ServerPort)
	app.Listen(fmt.Sprintf(":%v", config.ServerPort))
}
