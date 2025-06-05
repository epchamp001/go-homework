package mappers

import (
	"pvz-cli/pkg/errs"
	pvzpb "pvz-cli/pkg/pvz"
	"strconv"
)

func DomainProcessResultToProtoProcessResult(result map[string]error) (*pvzpb.ProcessResult, error) {
	if result == nil {
		return &pvzpb.ProcessResult{}, nil
	}

	processed := make([]uint64, 0, len(result))
	errorsList := make([]uint64, 0, len(result))

	for idStr, entryErr := range result {
		idUint, err := strconv.ParseUint(idStr, 10, 64)
		if err != nil {
			return nil, errs.Wrap(err, errs.CodeParsingError, "invalid key in ProcessResult", "order_id", idStr)
		}
		if entryErr != nil {
			errorsList = append(errorsList, idUint)
		} else {
			processed = append(processed, idUint)
		}
	}

	return &pvzpb.ProcessResult{
		Processed: processed,
		Errors:    errorsList,
	}, nil
}
