package server

import (
	"google.golang.org/api/drive/v3"
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
