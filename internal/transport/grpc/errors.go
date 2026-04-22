package grpctransport

import (
	"errors"

	"github.com/ljubushkin/container-management-service/internal/apperror"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func mapError(err error) error {
	var appErr *apperror.AppError

	if errors.As(err, &appErr) {
		switch appErr.Code {

		case apperror.CodeNotFound:
			return status.Error(codes.NotFound, appErr.Message)

		case apperror.CodeInvalidType,
			apperror.CodeInvalidStatus,
			apperror.CodeInvalidWarehouse,
			apperror.CodeInvalidPagination:
			return status.Error(codes.InvalidArgument, appErr.Message)

		default:
			return status.Error(codes.Internal, appErr.Message)
		}
	}

	return status.Error(codes.Internal, "internal error")
}
