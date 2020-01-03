package main

import (
	client "github.com/aiven/aiven-go-client"
	"log"
	"os"
	"time"
)

func main() {
	// Create new user client
	c, err := client.NewUserClient(
		os.Getenv("AIVEN_USERNAME"),
		os.Getenv("AIVEN_PASSWORD"), "aiven-go-client-test/"+client.Version())
	if err != nil {
		log.Fatalf("user authentication error: %s", err)
	}

	// Create new project
	project, err := c.Projects.Create(client.CreateProjectRequest{
		CardID:  os.Getenv("AIVEN_CARD_ID"),
		Cloud:   "google-europe-west1",
		Project: "testproject2",
	})
	if err != nil {
		log.Fatalf("project creation error: %s", err)
	}

	// Create new Kafka service inside the project
	userConfig := make(map[string]interface{})
	userConfig["kafka_version"] = "2.4"
	userConfig["schema_registry"] = true

	service, err := c.Services.Create(project.Name, client.CreateServiceRequest{
		Cloud:                 "google-europe-west1",
		GroupName:             "default",
		MaintenanceWindow:     nil,
		Plan:                  "business-4",
		ProjectVPCID:          nil,
		ServiceName:           "kafka-test-service",
		ServiceType:           "kafka",
		TerminationProtection: false,
		UserConfig:            userConfig,
		ServiceIntegrations:   nil,
	})
	if err != nil {
		log.Fatalf("cannot create new Kafka service, error: %s", err)
	}

	for {
		schema, err := c.KafkaSchemas.AddSubject(project.Name, service.Name, "test1", client.NewKafkaSchema(
			`
			{
				"doc": "example",
				"fields": [{
					"default": 5,
					"doc": "my test number",
					"name": "test",
					"namespace": "test",
					"type": "int"
				}],
				"name": "example",
				"namespace": "example",
				"type": "record"
			}
		`))

		if err != nil {
			//service is not started yet, and creation of a new ACL is not available yet
			if err.(client.Error).Status == 501 {
				log.Print("Kafka service is not started yet, err :" + err.Error())
				log.Print("Next attempt after 10 seconds delay ...")
				time.Sleep(10 * time.Second)
				continue
			}

			log.Fatalf("cannot create new Kafka Schema, error: %s", err)
		}

		log.Printf("Kafka Schema created, id %d", schema.Id)
		break
	}

	_, err = c.KafkaSchemas.UpdateConfig(project.Name, service.Name, client.KafkaSchemaConfig{CompatibilityLevel: "FULL"})
	if err != nil {
		log.Fatalf("cannot update Kafka Schema Configuration, error: %s", err)
	}
}
