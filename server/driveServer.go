package server

import (
	"context"
	"io/ioutil"

	"golang.org/x/oauth2/google"
	"google.golang.org/api/docs/v1"
	"google.golang.org/api/drive/v3"
	"google.golang.org/api/option"
)

/*
Encapsulate all interactions with Google Drive API
  - Auth
  - Load metadata/payload
  - Update metadata/payload
*/
type DriveServer struct {
	driveService     *drive.driveService
	docsService      *docs.Service
	shardDocumentIDs map[string]string // shardID: gDocID
}

func NewDriveServer(driveService *drive.Service, docsService *docs.Service) *DriveClient {

}

func NewDriveServerFromSecret(secretPath string) (*DriveServer, error) {
	ctx := context.Background()
	credsData, err := ioutil.ReadFile(secretPath)
	if err != nil {
		return nil, err
	}

	config, err := google.JWTConfigFromJSON(credsData, drive.DriveScope, docs.DocumentsScope)
	if err != nil {
		return nil, err
	}

	client := config.Client(ctx)
	driveService, err := drive.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		return nil, err
	}
	docsService, err := docs.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		return nil, err
	}

	return &DriveServer{
		driveService:     driveService,
		docsService:      docsService,
		shardDocumentIDs: make(map[string]string),
	}, nil
}
