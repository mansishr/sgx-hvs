/*
 * Copyright (C) 2019 Intel Corporation
 * SPDX-License-Identifier: BSD-3-Clause
 */
package resource

import (
	"encoding/json"
	"net/http"
	"time"

	//	"fmt"
	uuid "github.com/google/uuid"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/pkg/errors"
	"strconv"

	"intel/isecl/lib/common/v2/validation"
	"intel/isecl/sgx-host-verification-service/constants"
	"intel/isecl/sgx-host-verification-service/repository"
	"intel/isecl/sgx-host-verification-service/types"
)

type ResponseJson struct {
	Id      string
	Status  string
	Message string
}

type RegisterResponse struct {
	Response   ResponseJson
	HttpStatus int
}

type RegisterHostInfo struct {
	HostId           string `json:"host_ID"`
	HostName         string `json:"host_name"`
	ConnectionString string `json:"connection_string"`
	Description      string `json:"description, omitempty"`
	UUID             string `json:"uuid"`
	Overwrite        bool   `json:"overwrite"`
}

type AttReportThreadData struct {
	Uuid string
	Conn repository.SHVSDatabase
}

func SGXHostRegisterOps(r *mux.Router, db repository.SHVSDatabase) {
	log.Trace("resource/registerhost_ops: RegisterHostOps() Entering")
	defer log.Trace("resource/registerhost_ops: RegisterHostOps() Leaving")

	r.Handle("/hosts", handlers.ContentTypeHandler(RegisterHostCB(db), "application/json")).Methods("POST")
	r.Handle("/hosts/{id}", handlers.ContentTypeHandler(GetHostsCB(db), "application/json")).Methods("GET")
	r.Handle("/hosts", handlers.ContentTypeHandler(QueryHostsCB(db), "application/json")).Methods("GET")
	r.Handle("/platform-data", handlers.ContentTypeHandler(GetPaltformDataCB(db), "application/json")).Methods("GET")
	r.Handle("/reports", handlers.ContentTypeHandler(RetriveHostAttestationReportCB(db), "application/json")).Methods("GET")
	r.Handle("/latestPerHost", handlers.ContentTypeHandler(RetriveHostAttestationReportCB(db), "application/json")).Methods("GET")
	r.Handle("/host-status", handlers.ContentTypeHandler(HostStateInformationCB(db), "application/json")).Methods("GET")
	r.Handle("/hosts/{id}", DeleteHostCB(db)).Methods("DELETE")
}

func GetHostsCB(db repository.SHVSDatabase) errorHandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) error {
		log.Info("GetHostsCB entering")
		id := mux.Vars(r)["id"]
		log.Info("id: ", id)
		validation_err := validation.ValidateUUIDv4(id)
		if validation_err != nil {
			return &resourceError{Message: validation_err.Error(), StatusCode: http.StatusBadRequest}
		}

		ext_host, err := db.HostRepository().Retrieve(types.Host{Id: id})
		if ext_host == nil || err != nil {
			log.WithError(err).WithField("id", id).Info("attempt to fetch invalid host")
			return &resourceError{Message: "Host with given id don't exist",
				StatusCode: http.StatusNotFound}
		}

		host_Info := RegisterHostInfo{
			HostId:           ext_host.Id,
			HostName:         ext_host.Name,
			ConnectionString: ext_host.ConnectionString,
			UUID:             ext_host.HardwareUUID,
		}

		///Write the output here.
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		js, err := json.Marshal(host_Info)
		log.Info("host information: ", host_Info)
		if err != nil {
			log.Debug("Marshalling unsuccessful")
			return &resourceError{Message: err.Error(), StatusCode: http.StatusInternalServerError}
		}
		w.Write(js)
		log.Info("GetHostsCB leaving")
		return nil
	}
}

func QueryHostsCB(db repository.SHVSDatabase) errorHandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) error {
		log.Debug("QueryHostsCB entering")

		log.Debug("query", r.URL.Query())
		hardwareUUID := r.URL.Query().Get("HardwareUUID")
		hostName := r.URL.Query().Get("HostName")

		if hostName != "" {
			if !ValidateInputString(constants.HostName, hostName) {
				return &resourceError{Message: "QueryHostsCB: Invalid query Param Data",
					StatusCode: http.StatusBadRequest}
			}
		}

		if hardwareUUID != "" {
			if !ValidateInputString(constants.UUID, hardwareUUID) {
				return &resourceError{Message: "QueryHostsCB: Invalid query Param Data",
					StatusCode: http.StatusBadRequest}
			}
		}

		filter := types.Host{
			HardwareUUID: hardwareUUID,
			Name:         hostName,
		}

		hostData, err := db.HostRepository().GetHostQuery(&filter)

		if err != nil {
			log.WithError(err).WithField("filter", filter).Info("failed to retrieve roles")
			return &resourceError{Message: err.Error(), StatusCode: http.StatusInternalServerError}
		}
		if len(hostData) == 0 {
			log.Info("no data is found")
			return &resourceError{Message: "no host is found", StatusCode: http.StatusOK}
		}
		log.Info("hostData: ", hostData)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK) // HTTP 200
		//js, err := json.Marshal(fmt.Sprintf("%s", hostData))
		js, err := json.Marshal(hostData)
		if err != nil {
			return &resourceError{Message: err.Error(), StatusCode: http.StatusInternalServerError}
		}
		w.Write(js)
		log.Debug("QueryHosts leaving")
		return nil
	}
}

func GetPaltformDataCB(db repository.SHVSDatabase) errorHandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) error {
		log.Debug("GetPaltformDataCB entering")

		log.Debug("query", r.URL.Query())
		var platformData types.HostsSgxData
		hostName := r.URL.Query().Get("HostName")
		if hostName != "" {
			if !ValidateInputString(constants.HostName, hostName) {
				return &resourceError{Message: "QueryHostsCB: Invalid query Param Data",
					StatusCode: http.StatusBadRequest}
			}
			rs := types.Host{Name: hostName}
			///Get hosts data with the given hostname
			hostData, err := db.HostRepository().Retrieve(rs)
			if err != nil {
				log.WithError(err).WithField("HostName", hostName).Info("failed to retrieve hosts")
				return &resourceError{Message: err.Error(), StatusCode: http.StatusInternalServerError}
			}
			rs1 := types.HostSgxData{HostId: hostData.Id}
			platformData, err = db.HostSgxDataRepository().RetrieveAll(rs1)
			if err != nil {
				log.WithError(err).WithField("HostName", hostName).Info("failed to retrieve host data")
				return &resourceError{Message: err.Error(), StatusCode: http.StatusInternalServerError}
			}

		} else {
			numberOfMinutes := r.URL.Query().Get("numberOfMinutes")
			if numberOfMinutes != "" {
				_, err := strconv.Atoi(numberOfMinutes)
				if err != nil {
					log.Info("error came: ", err)
					return &resourceError{Message: "GetPaltformDataCB : Invalid query Param Data",
						StatusCode: http.StatusBadRequest}
				}
			}
			///Get all the hosts from host_statuses who are updated recently and status="CONNECTED"
			m, _ := time.ParseDuration(numberOfMinutes + "m")

			updatedTime := time.Now().Add(time.Duration((-m)))

			var err error
			platformData, err = db.HostSgxDataRepository().GetPlatformData(updatedTime)
			if err != nil {
				log.WithError(err).WithField("numberOfMinutes", updatedTime).Info("failed to retrieve updated hosts")
				return &resourceError{Message: err.Error(), StatusCode: http.StatusInternalServerError}
			}
		}
		log.Info("platformData: ", platformData)
		if len(platformData) == 0 {
			log.Info("no data is updated")
			return &resourceError{Message: "no host is updated", StatusCode: http.StatusOK}
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK) // HTTP 200
		js, err := json.Marshal(platformData)
		if err != nil {
			return &resourceError{Message: err.Error(), StatusCode: http.StatusInternalServerError}
		}
		w.Write(js)

		log.Debug("GetPaltformDataCB leaving")
		return nil
	}
}

func DeleteHostCB(db repository.SHVSDatabase) errorHandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) error {
		id := mux.Vars(r)["id"]
		validation_err := validation.ValidateUUIDv4(id)
		if validation_err != nil {
			return &resourceError{Message: validation_err.Error(), StatusCode: http.StatusBadRequest}
		}

		ext_host, err := db.HostRepository().Retrieve(types.Host{Id: id})
		if ext_host == nil || err != nil {
			log.WithError(err).WithField("id", id).Info("attempt to delete invalid host")
			w.WriteHeader(http.StatusNoContent)
			return nil
		}

		host := types.Host{
			Id:               ext_host.Id,
			Name:             ext_host.Name,
			Description:      ext_host.Description,
			ConnectionString: ext_host.ConnectionString,
			HardwareUUID:     ext_host.HardwareUUID,
			CreatedTime:      ext_host.CreatedTime,
			UpdatedTime:      time.Now(),
			Deleted:          true,
		}
		err = db.HostRepository().Update(host)
		if err != nil {
			return errors.New("DeleteHostCB: Error while Updating Host Information: " + err.Error())
		}
		slog.WithField("user", ext_host).Info("User deleted by:", r.RemoteAddr)
		err = UpdateHostStatus(ext_host.Id, db, constants.HostStatusRemoved)
		if err != nil {
			return errors.New("DeleteHostCB: Error while Updating Host Status Information: " + err.Error())
		}
		w.WriteHeader(http.StatusNoContent)
		return nil
	}
}

func UpdateSGXHostInfo(db repository.SHVSDatabase, existingHostData *types.Host, hostInfo RegisterHostInfo) error {
	log.Debug("UpdateSGXHostInfo: caching sgx data:", hostInfo)

	host := types.Host{
		Id:               existingHostData.Id,
		Name:             hostInfo.HostName,
		Description:      hostInfo.Description,
		ConnectionString: hostInfo.ConnectionString,
		HardwareUUID:     hostInfo.UUID,
		CreatedTime:      existingHostData.CreatedTime,
		UpdatedTime:      time.Now(),

		Deleted: false,
	}
	err := db.HostRepository().Update(host)
	if err != nil {
		return errors.New("UpdateSGXHostInfo: Error while Updating Host Information: " + err.Error())
	}

	err = UpdateHostStatus(existingHostData.Id, db, constants.HostStatusAgentQueued)
	if err != nil {
		log.Debug("UpdateSGXHostInfo failed")
		return errors.New("UpdateSGXHostInfo: Error while Updating Host Status Information: " + err.Error())
	}
	log.Debug("UpdateSGXHostInfo: Update SGX Host Data")
	return nil
}

func CreateSGXHostInfo(db repository.SHVSDatabase, hostInfo RegisterHostInfo) (string, error) {
	log.Debug("CreateSGXHostInfo: caching sgx data:", hostInfo)

	hostId := uuid.New().String()
	host := types.Host{
		Id:               hostId,
		Name:             hostInfo.HostName,
		Description:      hostInfo.Description,
		ConnectionString: hostInfo.ConnectionString,
		HardwareUUID:     hostInfo.UUID,
		CreatedTime:      time.Now(),
		UpdatedTime:      time.Now(),
	}
	_, err := db.HostRepository().Create(host)
	if err != nil {
		return "", errors.New("CreateSGXHostInfo: Error while caching Host Information: " + err.Error())
	}

	hostStatus := types.HostStatus{
		Id:          uuid.New().String(),
		HostId:      hostId,
		Status:      constants.HostStatusAgentQueued,
		CreatedTime: time.Now(),
		UpdatedTime: time.Now(),
	}

	_, err = db.HostStatusRepository().Create(hostStatus)
	if err != nil {
		return "", errors.New("CreateSGXHostInfo: Error while caching Host Status Information: " + err.Error())
	}

	log.Debug("CreateSGXHostInfo: Insert SGX Host Data")
	return hostId, nil
}

func SendHostRegisterResponse(w http.ResponseWriter, res RegisterResponse) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(res.HttpStatus)

	js, err := json.Marshal(res.Response)
	if err != nil {
		return errors.New("SendHostRegisterResponse: " + err.Error())
	}
	w.Write(js)
	return nil
}

func RegisterHostCB(db repository.SHVSDatabase) errorHandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) error {

		var res RegisterResponse
		var data RegisterHostInfo
		if r.ContentLength == 0 {
			res = RegisterResponse{HttpStatus: http.StatusBadRequest,
				Response: ResponseJson{Status: "Failed",
					Message: "RegisterHostCB: No request data"}}
			return SendHostRegisterResponse(w, res)
		}

		dec := json.NewDecoder(r.Body)
		dec.DisallowUnknownFields()
		err := dec.Decode(&data)
		if err != nil {
			res = RegisterResponse{HttpStatus: http.StatusBadRequest,
				Response: ResponseJson{Status: "Failed",
					Message: "RegisterHostCB: Invalid Json Post Data"}}
			return SendHostRegisterResponse(w, res)
		}

		log.Debug("Calling RegisterHostCB.................", data)

		if !ValidateInputString(constants.HostName, data.HostName) ||
			!ValidateInputString(constants.ConnectionString, data.ConnectionString) ||
			!ValidateInputString(constants.UUID, data.UUID) ||
			!ValidateInputString(constants.Description, data.Description) {

			res = RegisterResponse{HttpStatus: http.StatusBadRequest,
				Response: ResponseJson{Status: "Failed",
					Message: "RegisterHostCB: Invalid query Param Data"}}
			return SendHostRegisterResponse(w, res)
		}

		host := &types.Host{
			HardwareUUID: data.UUID,
		}

		existingHostData, err := db.HostRepository().Retrieve(*host)
		if existingHostData != nil && data.Overwrite == false {
			res = RegisterResponse{HttpStatus: http.StatusOK,
				Response: ResponseJson{Status: "Success",
					Id:      existingHostData.Id,
					Message: "Host already registerd in SGX HVS"}}
			return SendHostRegisterResponse(w, res)
		} else if existingHostData != nil && data.Overwrite == true {
			err = UpdateSGXHostInfo(db, existingHostData, data)
			if err != nil {
				res = RegisterResponse{HttpStatus: http.StatusInternalServerError,
					Response: ResponseJson{Status: "Failed",
						Message: "RegisterHostCB: " + err.Error()}}
				return SendHostRegisterResponse(w, res)
			}

			res = RegisterResponse{HttpStatus: http.StatusCreated,
				Response: ResponseJson{Status: "Created",
					Id:      existingHostData.Id,
					Message: "SGX Host Re-registered Successfully"}}
			return SendHostRegisterResponse(w, res)

		} else if existingHostData == nil { //if existingHostData == nil

			hostId, err := CreateSGXHostInfo(db, data)
			if err != nil {
				res = RegisterResponse{HttpStatus: http.StatusInternalServerError,
					Response: ResponseJson{Status: "Failed",
						Message: "RegisterHostCB: " + err.Error()}}
				return SendHostRegisterResponse(w, res)
			}

			res = RegisterResponse{HttpStatus: http.StatusCreated,
				Response: ResponseJson{Status: "Success",
					Id:      hostId,
					Message: "SGX Host Registered Successfully"}}
			return SendHostRegisterResponse(w, res)
		} else {

			res = RegisterResponse{HttpStatus: http.StatusInternalServerError,
				Response: ResponseJson{Status: "Failed",
					Message: "Invalid data"}}
			return SendHostRegisterResponse(w, res)
		}
	}
}