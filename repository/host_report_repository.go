/*
 * Copyright (C) 2019 Intel Corporation
 * SPDX-License-Identifier: BSD-3-Clause
 */
package repository

import (
	"intel/isecl/sgx-host-verification-service/types"
)

type HostReportRepository interface {
	Create(types.HostReport) (*types.HostReport, error)
	Retrieve(types.HostReport) (*types.HostReport, error)
	RetrieveAll(user types.HostReport) (types.HostReports, error)
	Update(types.HostReport) error
	Delete(types.HostReport) error
	GetHostReportQuary(types.SgxHostReportInputData) (types.HostReports, error)
}