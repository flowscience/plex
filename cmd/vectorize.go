package cmd

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"

	"github.com/google/uuid"
	"github.com/labdao/plex/internal/ipfs"
	"github.com/spf13/cobra"
)

var (
	cidOnly bool
)

var vectorizeCmd = &cobra.Command{
	Use:   "upload",
	Short: "Upload a file or directory to IPFS",
	Long:  `Upload and pins a file or directory to IPFS. Will wrap single files in a directory before uploading. (50MB max). A`,
	Run: func(cmd *cobra.Command, args []string) {
	},
}

func VectorizeOutputs(ioPath string, toolCid string, outputDir string) (map[string]OutputValues, error) {
	isCID := ipfs.IsValidCID(ioPath)
	id := uuid.New()

	cwd, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	if outputDir != "" {
		absPath, err := filepath.Abs(outputDir)
		if err != nil {
			return nil, err
		}
		cwd = absPath
	}

	workDirPath := path.Join(cwd, id.String())
	err = os.Mkdir(workDirPath, 0755)
	if err != nil {
		return nil, err
	}

	if isCID {
		ioPath = ipfs.DownloadToDirectory(ioPath, workDirPath)
	} else {
		ioPath, err = filepath.Abs(ioPath)
		if err != nil {
			return nil, err
		}
	}

	file, err := os.Open(ioPath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	bytes, err := ioutil.ReadAll(file)
	if err != nil {
		return nil, err
	}

	var ios []IO
	err = json.Unmarshal(bytes, &ios)
	if err != nil {
		return nil, err
	}

	outputMap := make(map[string]OutputValues)
	for i, io := range ios {
		if io.Tool.IPFS == toolCid {
			for key, output := range io.Outputs {
				fileOutput, ok := output.(FileOutput)
				if ok {
					ov := outputMap[key]

					filePath := fmt.Sprintf("entry-%d/outputs/%s", i, fileOutput.FilePath)
					absoluteFilePath := path.Join(workDirPath, filePath)

					// Download the file from IPFS to the local file path
					ipfs.DownloadToDirectory(fileOutput.IPFS, filepath.Dir(absoluteFilePath))

					ov.FilePaths = append(ov.FilePaths, absoluteFilePath)
					ov.CIDs = append(ov.CIDs, fileOutput.IPFS)
					outputMap[key] = ov
				}
			}
		}
	}

	if isCID {
		// clean up local files after use
		os.Remove(ioPath)
	}

	return outputMap, nil
}

// if path is CID and cid only is true
// download iojson and vectorize along CIDs (raise not implemented error)
// if path is CID and cid only is false
// create/dl working directory and vectorize along CIDs and local filepaths (raise not implemented error)
// if path is local file and cid only is true
// read file and vectorize along CIDs (raise not implemented error)
// if path is io json
// go thru each entry and
// return

func init() {
	vectorizeCmd.Flags().StringVarP(&toolPath, "toolPath", "p", "", "CID or file path of Tool config")
	vectorizeCmd.Flags().BoolVarP(&cidOnly, "cidOnly", "", false, "Only vectorize output CIDs")

	rootCmd.AddCommand(uploadCmd)
}
