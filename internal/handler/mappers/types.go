package mappers

import (
	"pvz-cli/internal/domain/models"
	pvzpb "pvz-cli/pkg/pvz"
)

func ProtoToDomainPackage(p pvzpb.PackageType) models.PackageType {
	switch p {
	case pvzpb.PackageType_PACKAGE_TYPE_BAG:
		return models.PackageBag
	case pvzpb.PackageType_PACKAGE_TYPE_BOX:
		return models.PackageBox
	case pvzpb.PackageType_PACKAGE_TYPE_TAPE:
		return models.PackageFilm
	case pvzpb.PackageType_PACKAGE_TYPE_BAG_TAPE:
		return models.PackageBagFilm
	case pvzpb.PackageType_PACKAGE_TYPE_BOX_TAPE:
		return models.PackageBoxFilm
	default:
		return models.PackageNone
	}
}

func DomainToProtoPackage(p models.PackageType) pvzpb.PackageType {
	switch p {
	case models.PackageBag:
		return pvzpb.PackageType_PACKAGE_TYPE_BAG
	case models.PackageBox:
		return pvzpb.PackageType_PACKAGE_TYPE_BOX
	case models.PackageFilm:
		return pvzpb.PackageType_PACKAGE_TYPE_TAPE
	case models.PackageBagFilm:
		return pvzpb.PackageType_PACKAGE_TYPE_BAG_TAPE
	case models.PackageBoxFilm:
		return pvzpb.PackageType_PACKAGE_TYPE_BOX_TAPE
	default:
		return pvzpb.PackageType_PACKAGE_TYPE_UNSPECIFIED
	}
}

func ProtoToDomainOrderStatus(s pvzpb.OrderStatus) models.OrderStatus {
	switch s {
	case pvzpb.OrderStatus_ORDER_STATUS_EXPECTS:
		return models.StatusAccepted
	case pvzpb.OrderStatus_ORDER_STATUS_ACCEPTED:
		return models.StatusIssued
	case pvzpb.OrderStatus_ORDER_STATUS_RETURNED:
		return models.StatusReturned
	case pvzpb.OrderStatus_ORDER_STATUS_DELETED:
		return models.StatusExpired
	default:
		return ""
	}
}

func DomainToProtoOrderStatus(s models.OrderStatus) pvzpb.OrderStatus {
	switch s {
	case models.StatusAccepted:
		return pvzpb.OrderStatus_ORDER_STATUS_EXPECTS
	case models.StatusIssued:
		return pvzpb.OrderStatus_ORDER_STATUS_ACCEPTED
	case models.StatusReturned:
		return pvzpb.OrderStatus_ORDER_STATUS_RETURNED
	case models.StatusExpired:
		return pvzpb.OrderStatus_ORDER_STATUS_DELETED
	default:
		return pvzpb.OrderStatus_ORDER_STATUS_UNSPECIFIED
	}
}
