package pd

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/imroc/req"
	"github.com/itsDarianNgo/go-pd/pkg/pd/utils"
)

const (
	Name             = "PixelDrain.com"
	BaseURL          = "https://pixeldrain.com/"
	APIURL           = BaseURL + "api"
	DefaultUserAgent = "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/79.0.3945.117 Safari/537.36"
	// errors
	ErrMissingPathToFile = "file path or file reader is required"
	ErrMissingFileID     = "file id is required"
	ErrMissingFilename   = "if you use ReadCloser you need to specify the filename"
	CSVFilePath          = "upload_logs.csv" // Path to the CSV file
)

type ClientOptions struct {
	Debug             bool
	ProxyURL          string
	EnableCookies     bool
	EnableInsecureTLS bool
	Timeout           time.Duration
}

type Client struct {
	Header  req.Header
	Request *req.Req
}

type PixelDrainClient struct {
	Client *Client
	Debug  bool
}

// New - create a new PixelDrainClient
func New(opt *ClientOptions, c *Client) *PixelDrainClient {
	// set default values if no other options available
	if opt == nil {
		opt = &ClientOptions{
			Debug:             false,
			ProxyURL:          "",
			EnableCookies:     true,
			EnableInsecureTLS: true,
			Timeout:           1 * time.Hour,
		}
	}

	// build default client if not available
	if c == nil {
		c = &Client{
			Header: req.Header{
				"User-Agent": DefaultUserAgent,
			},
			Request: req.New(),
		}
	}

	// set the request options
	c.Request.EnableCookie(opt.EnableCookies)
	c.Request.EnableInsecureTLS(opt.EnableInsecureTLS)
	c.Request.SetTimeout(opt.Timeout)
	if opt.ProxyURL != "" {
		_ = c.Request.SetProxyUrl(opt.ProxyURL)
	}

	pdc := &PixelDrainClient{
		Client: c,
		Debug:  opt.Debug,
	}

	return pdc
}

// UploadPOST POST /api/file | Updated method to include directory upload functionality
// curl -X POST -i -H "Authorization: Basic <TOKEN>" -F "file=@cat.jpg" https://pixeldrain.com/api/file
func (pd *PixelDrainClient) UploadPOST(r *RequestUpload, hashFilePath string) (*ResponseUpload, error) {
	if r.PathToFile == "" && r.File == nil {
		return nil, errors.New(ErrMissingPathToFile)
	}

	// Check if PathToFile is a directory
	if r.PathToFile != "" {
		fileInfo, err := os.Stat(r.PathToFile)
		if err != nil {
			return nil, err
		}
		if fileInfo.IsDir() {
			// If it's a directory, use UploadDirectory method
			return nil, pd.UploadDirectory(r.PathToFile, r.Auth, hashFilePath)
		}
	}

	// Check for duplicate file
	if r.PathToFile != "" {
		isDuplicate, err := utils.IsDuplicate(hashFilePath, r.PathToFile)
		if err != nil {
			return nil, err
		}
		if isDuplicate {
			log.Printf("File %s is a duplicate. Skipping upload.", r.PathToFile)
			return &ResponseUpload{
				ResponseDefault: ResponseDefault{
					Success:    false,
					StatusCode: http.StatusConflict,
					Message:    "Duplicate file. Upload skipped.",
				},
			}, nil
		}
	}

	return pd.uploadFile(r, hashFilePath)
}

func (pd *PixelDrainClient) uploadFile(r *RequestUpload, hashFilePath string) (*ResponseUpload, error) {
	if r.URL == "" {
		r.URL = fmt.Sprint(APIURL + "/file")
	}

	reqFileUpload := req.FileUpload{}
	var filePath string
	var fileSize int64
	var mimeType string

	log.Printf("Starting upload for file: %s", r.PathToFile)
	if r.File != nil {
		if r.FileName == "" {
			return nil, errors.New(ErrMissingFilename)
		}
		reqFileUpload.FileName = r.FileName
		reqFileUpload.FieldName = "file"

		// Read the file into a buffer to determine the MIME type and size
		var buf bytes.Buffer
		size, err := io.Copy(&buf, r.File)
		if err != nil {
			return nil, err
		}
		r.File.Close()              // Close the original ReadCloser
		r.File = io.NopCloser(&buf) // Reset the file reader

		mimeType = http.DetectContentType(buf.Bytes()[:512])
		fileSize = size
		reqFileUpload.File = io.NopCloser(bytes.NewReader(buf.Bytes()))

		// Attempt to use the PathToFile if provided, otherwise mark as "N/A"
		if r.PathToFile != "" {
			filePath = r.PathToFile
		} else {
			filePath = "N/A" // No file path when using io.ReadCloser
		}
	} else {
		file, err := os.Open(r.PathToFile)
		if err != nil {
			return nil, err
		}
		defer func() {
			if cerr := file.Close(); cerr != nil {
				log.Printf("Error closing file: %v", cerr)
			}
		}()

		reqFileUpload.FileName = filepath.Base(r.PathToFile)
		reqFileUpload.FieldName = "file"
		reqFileUpload.File = file

		filePath = r.PathToFile
		fileSize = utils.GetFileSize(filePath)
		mimeType = utils.GetMimeType(filePath)
	}

	reqParams := req.Param{
		"anonymous": r.Anonymous,
	}

	log.Printf("Sending POST request to %s with file: %s", r.URL, reqFileUpload.FileName)
	if r.Auth.IsAuthAvailable() && !r.Anonymous {
		addBasicAuthHeader(pd.Client.Header, "", r.Auth.APIKey)
	}

	rsp, err := pd.Client.Request.Post(r.URL, pd.Client.Header, reqFileUpload, reqParams)
	if pd.Debug {
		log.Println(rsp.Dump())
	}
	if err != nil {
		return nil, err
	}

	uploadRsp := &ResponseUpload{}
	uploadRsp.StatusCode = rsp.Response().StatusCode
	err = rsp.ToJSON(uploadRsp)
	if err != nil {
		log.Printf("Error parsing JSON response: %v", err)
		return nil, err
	}

	log.Printf("File uploaded successfully: %s", reqFileUpload.FileName)
	formattedFileSize := utils.FormatFileSize(fileSize)

	// Gather upload information and save it to CSV
	if filePath != "N/A" {
		uploadInfo := utils.UploadInfo{
			FileName:       reqFileUpload.FileName,
			DirectoryPath:  filePath,
			URL:            uploadRsp.GetFileURL(),
			UploadDateTime: time.Now().Format(time.RFC3339),
			FileSize:       fileSize,
			MIMEType:       mimeType,
			Uploader:       r.Auth.APIKey,
			UploadStatus:   fmt.Sprintf("%d", uploadRsp.StatusCode),
			FormattedSize:  formattedFileSize,
		}

		log.Printf("Logging upload info for file in uploadFile: %s", filePath)

		if err := utils.SaveUploadInfoToCSV(uploadInfo, CSVFilePath); err != nil {
			return nil, err
		}

		// Calculate the hash and save it to CSV
		fileHash, err := utils.CalculateFileHash(filePath)
		if err != nil {
			return nil, err
		}

		if err := utils.SaveFileHash(hashFilePath, filePath, fileHash); err != nil {
			return nil, err
		}
	}

	return uploadRsp, nil
}

// UploadPUT PUT /api/file/{name}
// curl -X PUT -i -H "Authorization: Basic <TOKEN>" --upload-file cat.jpg https://pixeldrain.com/api/file/test_cat.jpg
func (pd *PixelDrainClient) UploadPUT(r *RequestUpload) (*ResponseUpload, error) {
	if r.PathToFile == "" && r.File == nil {
		return nil, errors.New(ErrMissingPathToFile)
	}

	if r.File == nil && r.FileName == "" {
		return nil, errors.New(ErrMissingFilename)
	}

	if r.URL == "" {
		r.URL = fmt.Sprintf(APIURL+"/file/%s", r.GetFileName())
	}

	var file io.ReadCloser
	var err error
	if r.File != nil {
		file = r.File
	} else {
		file, err = os.Open(r.PathToFile)
		if err != nil {
			return nil, err
		}
	}

	// we don't send this parameter due a bug of pixeldrain side
	//reqParams := req.Param{
	//	"anonymous": r.Anonymous,
	//}

	// pixeldrain want an empty username and the APIKey as password
	if r.Auth.IsAuthAvailable() && !r.Anonymous {
		addBasicAuthHeader(pd.Client.Header, "", r.Auth.APIKey)
	}

	rsp, err := pd.Client.Request.Put(r.URL, pd.Client.Header, file)
	if pd.Debug {
		log.Println(rsp.Dump())
	}
	if err != nil {
		return nil, err
	}

	uploadRsp := &ResponseUpload{}
	uploadRsp.StatusCode = rsp.Response().StatusCode
	if uploadRsp.StatusCode == http.StatusCreated {
		uploadRsp.Success = true
	}
	err = rsp.ToJSON(uploadRsp)
	if err != nil {
		return nil, err
	}

	return uploadRsp, nil
}

// Download GET /api/file/{id}
func (pd *PixelDrainClient) Download(r *RequestDownload) (*ResponseDownload, error) {
	if r.PathToSave == "" {
		return nil, errors.New(ErrMissingPathToFile)
	}

	if r.ID == "" {
		return nil, errors.New(ErrMissingFileID)
	}

	if r.URL == "" {
		r.URL = fmt.Sprintf(APIURL+"/file/%s", r.ID)
	}

	// pixeldrain want an empty username and the APIKey as password
	if r.Auth.IsAuthAvailable() {
		addBasicAuthHeader(pd.Client.Header, "", r.Auth.APIKey)
	}

	rsp, err := pd.Client.Request.Get(r.URL, pd.Client.Header)
	if pd.Debug {
		log.Println(rsp.Dump())
	}
	if err != nil {
		return nil, err
	}

	if rsp.Response().StatusCode != 200 {
		defaultRsp := &ResponseDefault{}
		err = rsp.ToJSON(defaultRsp)
		if err != nil {
			return nil, err
		}

		defaultRsp.StatusCode = rsp.Response().StatusCode
		defaultRsp.Success = false

		downloadRsp := &ResponseDownload{
			ResponseDefault: *defaultRsp,
		}

		return downloadRsp, nil
	}

	err = rsp.ToFile(r.PathToSave)
	if err != nil {
		return nil, err
	}

	fInfo, err := os.Stat(r.PathToSave)
	if err != nil {
		return nil, err
	}

	downloadRsp := &ResponseDownload{
		FilePath: r.PathToSave,
		FileName: fInfo.Name(),
		FileSize: fInfo.Size(),
		ResponseDefault: ResponseDefault{
			StatusCode: rsp.Response().StatusCode,
			Success:    true,
		},
	}

	return downloadRsp, nil
}

// GetFileInfo GET /api/file/{id}/info
func (pd *PixelDrainClient) GetFileInfo(r *RequestFileInfo) (*ResponseFileInfo, error) {
	if r.ID == "" {
		return nil, errors.New(ErrMissingFileID)
	}

	if r.URL == "" {
		r.URL = fmt.Sprintf(APIURL+"/file/%s/info", r.ID)
	}

	// pixeldrain want an empty username and the APIKey as password
	if r.Auth.IsAuthAvailable() {
		addBasicAuthHeader(pd.Client.Header, "", r.Auth.APIKey)
	}

	rsp, err := pd.Client.Request.Get(r.URL, pd.Client.Header)
	if pd.Debug {
		log.Println(rsp.Dump())
	}
	if err != nil {
		return nil, err
	}

	fileInfoRsp := &ResponseFileInfo{}
	fileInfoRsp.StatusCode = rsp.Response().StatusCode
	if fileInfoRsp.StatusCode == http.StatusOK {
		fileInfoRsp.Success = true
	}
	err = rsp.ToJSON(fileInfoRsp)
	if err != nil {
		return nil, err
	}

	return fileInfoRsp, nil
}

// DownloadThumbnail GET /api/file/{id}/thumbnail?width=x&height=x
func (pd *PixelDrainClient) DownloadThumbnail(r *RequestThumbnail) (*ResponseThumbnail, error) {
	if r.PathToSave == "" {
		return nil, errors.New(ErrMissingPathToFile)
	}

	if r.ID == "" {
		return nil, errors.New(ErrMissingFileID)
	}

	if r.URL == "" {
		r.URL = fmt.Sprintf(APIURL+"/file/%s/thumbnail", r.ID)
	}

	queryParams := req.QueryParam{}
	if r.Width != "" {
		queryParams["width"] = r.Width
	}
	if r.Height != "" {
		queryParams["height"] = r.Height
	}

	// pixeldrain want an empty username and the APIKey as password
	if r.Auth.IsAuthAvailable() {
		addBasicAuthHeader(pd.Client.Header, "", r.Auth.APIKey)
	}

	rsp, err := pd.Client.Request.Get(r.URL, pd.Client.Header, queryParams)
	if pd.Debug {
		log.Println(rsp.Dump())
	}
	if err != nil {
		return nil, err
	}

	err = rsp.ToFile(r.PathToSave)
	if err != nil {
		return nil, err
	}

	fInfo, err := os.Stat(r.PathToSave)
	if err != nil {
		return nil, err
	}

	rspStruct := &ResponseThumbnail{
		FilePath: r.PathToSave,
		FileName: fInfo.Name(),
		FileSize: fInfo.Size(),
		ResponseDefault: ResponseDefault{
			StatusCode: rsp.Response().StatusCode,
			Success:    true,
		},
	}

	return rspStruct, nil
}

// Delete DELETE /api/file/{id}
func (pd *PixelDrainClient) Delete(r *RequestDelete) (*ResponseDelete, error) {
	if r.ID == "" {
		return nil, errors.New(ErrMissingFileID)
	}

	if r.URL == "" {
		r.URL = fmt.Sprintf(APIURL+"/file/%s", r.ID)
	}

	// pixeldrain want an empty username and the APIKey as password
	if r.Auth.IsAuthAvailable() {
		addBasicAuthHeader(pd.Client.Header, "", r.Auth.APIKey)
	}

	rsp, err := pd.Client.Request.Delete(r.URL, pd.Client.Header)
	if pd.Debug {
		log.Println(rsp.Dump())
	}
	if err != nil {
		return nil, err
	}

	rspStruct := &ResponseDelete{}
	err = rsp.ToJSON(rspStruct)
	if err != nil {
		return nil, err
	}

	rspStruct.StatusCode = rsp.Response().StatusCode

	return rspStruct, nil
}

// CreateList POST /api/list
func (pd *PixelDrainClient) CreateList(r *RequestCreateList) (*ResponseCreateList, error) {
	if r.URL == "" {
		r.URL = APIURL + "/list"
	}

	// pixeldrain want an empty username and the APIKey as password
	if r.Auth.IsAuthAvailable() && !r.Anonymous {
		addBasicAuthHeader(pd.Client.Header, "", r.Auth.APIKey)
	}

	data, err := json.Marshal(r)

	rsp, err := pd.Client.Request.Post(r.URL, pd.Client.Header, data)
	if pd.Debug {
		log.Println(rsp.Dump())
	}
	if err != nil {
		return nil, err
	}

	rspStruct := &ResponseCreateList{}
	err = rsp.ToJSON(rspStruct)
	if err != nil {
		return nil, err
	}

	rspStruct.StatusCode = rsp.Response().StatusCode

	return rspStruct, nil
}

// GetList GET /api/list/{id}
func (pd *PixelDrainClient) GetList(r *RequestGetList) (*ResponseGetList, error) {
	if r.ID == "" {
		return nil, errors.New(ErrMissingFileID)
	}

	if r.URL == "" {
		r.URL = fmt.Sprintf(APIURL+"/list/%s", r.ID)
	}

	// pixeldrain want an empty username and the APIKey as password
	if r.Auth.IsAuthAvailable() {
		addBasicAuthHeader(pd.Client.Header, "", r.Auth.APIKey)
	}

	rsp, err := pd.Client.Request.Get(r.URL, pd.Client.Header)
	if pd.Debug {
		log.Println(rsp.Dump())
	}
	if err != nil {
		return nil, err
	}

	rspStruct := &ResponseGetList{}
	err = rsp.ToJSON(rspStruct)
	if err != nil {
		return nil, err
	}

	rspStruct.StatusCode = rsp.Response().StatusCode

	return rspStruct, nil
}

// GetUser GET /api/user
func (pd *PixelDrainClient) GetUser(r *RequestGetUser) (*ResponseGetUser, error) {
	if r.URL == "" {
		r.URL = APIURL + "/user"
	}

	// pixeldrain want an empty username and the APIKey as password
	if r.Auth.IsAuthAvailable() {
		addBasicAuthHeader(pd.Client.Header, "", r.Auth.APIKey)
	}

	rsp, err := pd.Client.Request.Get(r.URL, pd.Client.Header)
	if pd.Debug {
		log.Println(rsp.Dump())
	}
	if err != nil {
		return nil, err
	}

	rspStruct := &ResponseGetUser{}
	err = rsp.ToJSON(rspStruct)
	if err != nil {
		return nil, err
	}

	status := false
	if rsp.Response().StatusCode == http.StatusOK {
		status = true
	}

	rspStruct.Success = status
	rspStruct.StatusCode = rsp.Response().StatusCode

	return rspStruct, nil
}

// GetUserFiles GET /api/user/files
func (pd *PixelDrainClient) GetUserFiles(r *RequestGetUserFiles) (*ResponseGetUserFiles, error) {
	if r.URL == "" {
		r.URL = APIURL + "/user/files"
	}

	// pixeldrain want an empty username and the APIKey as password
	if r.Auth.IsAuthAvailable() {
		addBasicAuthHeader(pd.Client.Header, "", r.Auth.APIKey)
	}

	rsp, err := pd.Client.Request.Get(r.URL, pd.Client.Header)
	if pd.Debug {
		log.Println(rsp.Dump())
	}
	if err != nil {
		return nil, err
	}

	rspStruct := &ResponseGetUserFiles{}
	err = rsp.ToJSON(rspStruct)
	if err != nil {
		return nil, err
	}

	status := false
	if rsp.Response().StatusCode == http.StatusOK {
		status = true
	}

	rspStruct.Success = status
	rspStruct.StatusCode = rsp.Response().StatusCode

	return rspStruct, nil
}

// GetUserLists GET /api/user/lists
func (pd *PixelDrainClient) GetUserLists(r *RequestGetUserLists) (*ResponseGetUserLists, error) {
	if r.URL == "" {
		r.URL = APIURL + "/user/lists"
	}

	// pixeldrain want an empty username and the APIKey as password
	if r.Auth.IsAuthAvailable() {
		addBasicAuthHeader(pd.Client.Header, "", r.Auth.APIKey)
	}

	rsp, err := pd.Client.Request.Get(r.URL, pd.Client.Header)
	if pd.Debug {
		log.Println(rsp.Dump())
	}
	if err != nil {
		return nil, err
	}

	rspStruct := &ResponseGetUserLists{}
	err = rsp.ToJSON(rspStruct)
	if err != nil {
		return nil, err
	}

	status := false
	if rsp.Response().StatusCode == http.StatusOK {
		status = true
	}

	rspStruct.Success = status
	rspStruct.StatusCode = rsp.Response().StatusCode

	return rspStruct, nil
}

// pixeldrain want an empty username and the APIKey as password
// addBasicAuthHeader create a http basic auth header from username and password
func addBasicAuthHeader(h req.Header, u string, p string) *req.Header {
	h["Authorization"] = "Basic " + generateBasicAuthToken(u, p)
	return &h
}

// generateBasicAuthToken generate string for basic auth header
func generateBasicAuthToken(u string, p string) string {
	auth := u + ":" + p
	return base64.StdEncoding.EncodeToString([]byte(auth))
}

// UploadDirectory uploads all files in the given directory and its subdirectories
func (pd *PixelDrainClient) UploadDirectory(directoryPath string, auth Auth, baseURL ...string) error {
	// Use the provided base URL if present
	apiURL := APIURL
	if len(baseURL) > 0 {
		apiURL = baseURL[0]
	}

	files, err := utils.GetFilesInDirectory(directoryPath)
	if err != nil {
		return err
	}

	// Get the appropriate hash file path based on the environment
	hashFilePath := utils.GetHashFilePath()

	for _, filePath := range files {
		reqUpload := &RequestUpload{
			PathToFile: filePath,
			Anonymous:  false,
			Auth:       auth,
			URL:        apiURL + "/file",
		}

		log.Printf("Uploading file: %s", filePath)
		resp, err := pd.UploadPOST(reqUpload, hashFilePath)
		if err != nil {
			log.Printf("Error uploading file %s: %v", filePath, err)
			return err
		}

		log.Printf("Upload response for file %s: %+v", filePath, resp)
	}

	return nil
}
