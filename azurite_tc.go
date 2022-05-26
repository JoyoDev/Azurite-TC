package azurite_tc

import (
	"context"
	"errors"
	"io"
	"os"
	"strings"

	"github.com/docker/go-connections/nat"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"

	"github.com/Azure/azure-sdk-for-go/storage"
	log "github.com/sirupsen/logrus"
)

type AzuriteTC struct {
	azuriteID                 string
	storageAccountName        string
	storageAccountAccessToken string
}

func NewAzuriteTC(storageAccountName, storageAccountAccessToken string) *AzuriteTC {
	return &AzuriteTC{
		azuriteID:                 "",
		storageAccountName:        storageAccountName,
		storageAccountAccessToken: storageAccountAccessToken,
	}
}

func (ac *AzuriteTC) RunAzuriteContainer() {
	ctx := context.Background()
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		panic(err)
	}

	imageName := "mcr.microsoft.com/azure-storage/azurite"

	out, err := cli.ImagePull(ctx, imageName, types.ImagePullOptions{})
	if err != nil {
		panic(err)
	}
	defer out.Close()
	_, err = io.Copy(os.Stdout, out)
	if err != nil {
		panic(err)
	}

	hostConfig := &container.HostConfig{
		PortBindings: nat.PortMap{
			nat.Port("10000/tcp"): []nat.PortBinding{{HostIP: "0.0.0.0", HostPort: "10000"}},
			nat.Port("10001/tcp"): []nat.PortBinding{{HostIP: "0.0.0.0", HostPort: "10001"}},
			nat.Port("10002/tcp"): []nat.PortBinding{{HostIP: "0.0.0.0", HostPort: "10002"}},
		},
	}

	resp, err := cli.ContainerCreate(ctx, &container.Config{
		Image: imageName,
		ExposedPorts: nat.PortSet{
			nat.Port("10000/tcp"): {},
			nat.Port("10001/tcp"): {},
			nat.Port("10002/tcp"): {},
		},
	}, hostConfig, nil, nil, "")
	if err != nil {
		panic(err)
	}

	if err := cli.ContainerStart(ctx, resp.ID, types.ContainerStartOptions{}); err != nil {
		panic(err)
	}

	ac.azuriteID = resp.ID
	log.WithField("container ID", ac.azuriteID).Debug("Memorizing container ID")
}

func (ac *AzuriteTC) RemoveAzuriteContainer() {
	ctx := context.Background()
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		panic(err)
	}

	log.WithField("container ID", ac.azuriteID[:10]).Debug("Stopping azurite container")

	if err := cli.ContainerStop(ctx, ac.azuriteID, nil); err != nil {
		panic(err)
	}

	log.WithField("container ID", ac.azuriteID[:10]).Debug("Removing azurite container")

	if err := cli.ContainerRemove(ctx, ac.azuriteID, types.ContainerRemoveOptions{}); err != nil {
		panic(err)
	}

	log.Debug("Azurite container removed")
}

// Create table on azurite if it does not exists
func (az *AzuriteTC) CreateTable(tableName string) {
	client, err := storage.NewBasicClient(az.storageAccountName, az.storageAccountAccessToken)
	if err != nil {
		panic(err)
	}

	tableService := client.GetTableService()
	table := tableService.GetTableReference(tableName)
	err = table.Create(30, storage.FullMetadata, nil)
	if err != nil {
		if !strings.Contains(err.Error(), "TableAlreadyExists") {
			panic(err)
		}
	} else {
		log.WithField("table", tableName).Debug("Table created")
	}
}

// Update value in the table by passing table name. partition and row keys
func (ac *AzuriteTC) UpdateTableValue(tableName, partitionKey, rowKey, value string) {
	client, err := storage.NewBasicClient(ac.storageAccountName, ac.storageAccountAccessToken)

	if err != nil {
		panic(err)
	}
	tableService := client.GetTableService()
	table := tableService.GetTableReference(tableName)
	entity := table.GetEntityReference(partitionKey, rowKey)

	tableBatch := table.NewBatch()

	props := map[string]interface{}{
		"value": value,
	}
	entity.Properties = props

	tableBatch.InsertOrReplaceEntity(entity, false)
	err = tableBatch.ExecuteBatch()
	if err != nil {
		if err.Error() != "unexpected EOF" {
			panic(err)
		}
	}

	log.WithField("value", value).Debug("Table value")
}

// Retrieve value from table
func (az *AzuriteTC) GetTablevalue(tableName, partitionKey, rowKey string) (string, error) {
	client, err := storage.NewBasicClient(az.storageAccountName, az.storageAccountAccessToken)

	if err != nil {
		return "", err
	}

	tableService := client.GetTableService()
	table := tableService.GetTableReference(tableName)
	entity := table.GetEntityReference(partitionKey, rowKey)

	err = entity.Get(5000, storage.FullMetadata, &storage.GetEntityOptions{Select: []string{"value"}})

	if err != nil {
		return "", err
	}

	value := entity.Properties["value"].(string)
	if value == "" {
		return "", errors.New("invalid result from table")
	}

	log.WithField("value", value).Debug("Retrived value from table")
	return value, nil
}
