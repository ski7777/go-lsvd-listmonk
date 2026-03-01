package sharepointfile

import (
	"context"
	"errors"
	"log"

	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	graph "github.com/microsoftgraph/msgraph-sdk-go"
	"github.com/microsoftgraph/msgraph-sdk-go/drives"
	"github.com/microsoftgraph/msgraph-sdk-go/models"
	"github.com/thoas/go-funk"
)

func GetFile(tenantid, clientid, clientsecret, driveid, folderitemid string) (data []byte, err error) {
	cred, err := azidentity.NewClientSecretCredential(tenantid, clientid, clientsecret, nil)
	if err != nil {
		return
	}

	graphClient, err := graph.NewGraphServiceClientWithCredentials(
		cred,
		[]string{"https://graph.microsoft.com/.default"},
	)
	if err != nil {
		return
	}

	ctx := context.TODO()

	childrenRaw, err := graphClient.
		Drives().
		ByDriveId(driveid).
		Items().
		ByDriveItemId(folderitemid).
		Children().
		Get(
			ctx,
			&drives.ItemItemsItemChildrenRequestBuilderGetRequestConfiguration{
				QueryParameters: &drives.ItemItemsItemChildrenRequestBuilderGetQueryParameters{
					Orderby: []string{"lastModifiedDateTime desc"},
				},
			},
		)
	if err != nil {
		return
	}
	children := funk.Filter(childrenRaw.GetValue(), func(item models.DriveItemable) bool {
		return *item.GetFile().GetMimeType() == "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"
	}).([]models.DriveItemable)
	if len(children) == 0 {
		err = errors.New("no excel files found")
		return
	}
	log.Println("Using file ", *children[0].GetName())
	data, err = graphClient.
		Drives().
		ByDriveId(driveid).
		Items().
		ByDriveItemId(*children[0].GetId()).
		Content().
		Get(
			ctx,
			nil,
		)
	if err != nil {
		return
	}
	return
}
