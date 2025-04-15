package adaptor

import (
	"github.com/devtron-labs/common-lib/imageScan/bean"
	bean2 "github.com/devtron-labs/image-scanner/pkg/sql/bean"
	"github.com/devtron-labs/image-scanner/pkg/sql/repository"
	"time"
)

func GetImageScanExecutionHistory(imageScanEvent *bean.ImageScanEvent, scanEventJson []byte, time time.Time) *repository.ImageScanExecutionHistory {
	return &repository.ImageScanExecutionHistory{
		Image:              imageScanEvent.Image,
		ImageHash:          imageScanEvent.ImageDigest,
		ExecutionTime:      time,
		ExecutedBy:         imageScanEvent.UserId,
		SourceMetadataJson: string(scanEventJson),
		SourceType:         imageScanEvent.SourceType,
		SourceSubType:      imageScanEvent.SourceSubType,
	}
}

func GetScanToolExecutionHistoryMapping(executionHistoryModel *repository.ImageScanExecutionHistory, state bean2.ScanExecutionProcessState,
	errorMsg string, toolId int) *repository.ScanToolExecutionHistoryMapping {
	return &repository.ScanToolExecutionHistoryMapping{
		ImageScanExecutionHistoryId: executionHistoryModel.Id,
		ScanToolId:                  toolId,
		ExecutionStartTime:          executionHistoryModel.ExecutionTime,
		State:                       state,
		ErrorMessage:                errorMsg,
		AuditLog: repository.AuditLog{
			CreatedOn: executionHistoryModel.ExecutionTime,
			CreatedBy: int32(executionHistoryModel.ExecutedBy),
			UpdatedOn: executionHistoryModel.ExecutionTime,
			UpdatedBy: int32(executionHistoryModel.ExecutedBy),
		},
	}
}

func GetResourceScanExecutionResult(executionHistoryId, scanToolId int, scanDataJson string, format repository.ResourceScanFormat) *repository.ResourceScanResult {
	var resultTypes []int
	switch format {
	case repository.SbomResultSource:
		resultTypes = []int{repository.Vulnerabilities.ToInt()}
	case repository.CycloneDxSbom:
		resultTypes = []int{repository.Vulnerabilities.ToInt(), repository.License.ToInt()}
	default:
		resultTypes = []int{repository.Vulnerabilities.ToInt(), repository.License.ToInt(), repository.Config.ToInt(), repository.Secrets.ToInt()}

	}
	return &repository.ResourceScanResult{
		ImageScanExecutionHistoryId: executionHistoryId,
		ScanDataJson:                scanDataJson,
		Format:                      format,
		Types:                       resultTypes,
		ScanToolId:                  scanToolId,
	}
}
