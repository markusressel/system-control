package util

import (
	"encoding/json"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

const (
	DiskByIdPath = "/dev/disk/by-id/"
)

// Example smartctl JSON output structure for HDD:
//
//	{
//	 "json_format_version" : [ 1, 0 ],
//	 "smartctl" : {
//	   "version" : [ 7, 5 ],
//	   "pre_release" : false,
//	   "svn_revision" : "5714",
//	   "platform_info" : "x86_64-linux-6.17.8-arch1-1",
//	   "build_info" : "(local build)",
//	   "argv" : [ "smartctl", "-json", "-A", "/dev/sdb" ],
//	   "drive_database_version" : {
//	     "string" : "7.5/5706"
//	   },
//	   "exit_status" : 0
//	 },
//	 "local_time" : {
//	   "time_t" : 1767314623,
//	   "asctime" : "Fri Jan  2 01:43:43 2026 CET"
//	 },
//	 "device" : {
//	   "name" : "/dev/sdb",
//	   "info_name" : "/dev/sdb [SAT]",
//	   "type" : "sat",
//	   "protocol" : "ATA"
//	 },
//	 "ata_smart_attributes" : {
//	   "revision" : 1,
//	   "table" : [ {
//	     "id" : 5,
//	     "name" : "Reallocated_Sector_Ct",
//	     "value" : 100,
//	     "worst" : 100,
//	     "thresh" : 10,
//	     "when_failed" : "",
//	     "flags" : {
//	       "value" : 51,
//	       "string" : "PO--CK ",
//	       "prefailure" : true,
//	       "updated_online" : true,
//	       "performance" : false,
//	       "error_rate" : false,
//	       "event_count" : true,
//	       "auto_keep" : true
//	     },
//	     "raw" : {
//	       "value" : 0,
//	       "string" : "0"
//	     }
//	   }, {
//	     "id" : 9,
//	     "name" : "Power_On_Hours",
//	     "value" : 92,
//	     "worst" : 92,
//	     "thresh" : 0,
//	     "when_failed" : "",
//	     "flags" : {
//	       "value" : 50,
//	       "string" : "-O--CK ",
//	       "prefailure" : false,
//	       "updated_online" : true,
//	       "performance" : false,
//	       "error_rate" : false,
//	       "event_count" : true,
//	       "auto_keep" : true
//	     },
//	     "raw" : {
//	       "value" : 36682,
//	       "string" : "36682"
//	     }
//	   }, {
//	     "id" : 12,
//	     "name" : "Power_Cycle_Count",
//	     "value" : 99,
//	     "worst" : 99,
//	     "thresh" : 0,
//	     "when_failed" : "",
//	     "flags" : {
//	       "value" : 50,
//	       "string" : "-O--CK ",
//	       "prefailure" : false,
//	       "updated_online" : true,
//	       "performance" : false,
//	       "error_rate" : false,
//	       "event_count" : true,
//	       "auto_keep" : true
//	     },
//	     "raw" : {
//	       "value" : 354,
//	       "string" : "354"
//	     }
//	   }, {
//	     "id" : 177,
//	     "name" : "Wear_Leveling_Count",
//	     "value" : 30,
//	     "worst" : 30,
//	     "thresh" : 0,
//	     "when_failed" : "",
//	     "flags" : {
//	       "value" : 19,
//	       "string" : "PO--C- ",
//	       "prefailure" : true,
//	       "updated_online" : true,
//	       "performance" : false,
//	       "error_rate" : false,
//	       "event_count" : true,
//	       "auto_keep" : false
//	     },
//	     "raw" : {
//	       "value" : 1259,
//	       "string" : "1259"
//	     }
//	   }, {
//	     "id" : 179,
//	     "name" : "Used_Rsvd_Blk_Cnt_Tot",
//	     "value" : 100,
//	     "worst" : 100,
//	     "thresh" : 10,
//	     "when_failed" : "",
//	     "flags" : {
//	       "value" : 19,
//	       "string" : "PO--C- ",
//	       "prefailure" : true,
//	       "updated_online" : true,
//	       "performance" : false,
//	       "error_rate" : false,
//	       "event_count" : true,
//	       "auto_keep" : false
//	     },
//	     "raw" : {
//	       "value" : 0,
//	       "string" : "0"
//	     }
//	   }, {
//	     "id" : 181,
//	     "name" : "Program_Fail_Cnt_Total",
//	     "value" : 100,
//	     "worst" : 100,
//	     "thresh" : 10,
//	     "when_failed" : "",
//	     "flags" : {
//	       "value" : 50,
//	       "string" : "-O--CK ",
//	       "prefailure" : false,
//	       "updated_online" : true,
//	       "performance" : false,
//	       "error_rate" : false,
//	       "event_count" : true,
//	       "auto_keep" : true
//	     },
//	     "raw" : {
//	       "value" : 0,
//	       "string" : "0"
//	     }
//	   }, {
//	     "id" : 182,
//	     "name" : "Erase_Fail_Count_Total",
//	     "value" : 100,
//	     "worst" : 100,
//	     "thresh" : 10,
//	     "when_failed" : "",
//	     "flags" : {
//	       "value" : 50,
//	       "string" : "-O--CK ",
//	       "prefailure" : false,
//	       "updated_online" : true,
//	       "performance" : false,
//	       "error_rate" : false,
//	       "event_count" : true,
//	       "auto_keep" : true
//	     },
//	     "raw" : {
//	       "value" : 0,
//	       "string" : "0"
//	     }
//	   }, {
//	     "id" : 183,
//	     "name" : "Runtime_Bad_Block",
//	     "value" : 100,
//	     "worst" : 100,
//	     "thresh" : 10,
//	     "when_failed" : "",
//	     "flags" : {
//	       "value" : 19,
//	       "string" : "PO--C- ",
//	       "prefailure" : true,
//	       "updated_online" : true,
//	       "performance" : false,
//	       "error_rate" : false,
//	       "event_count" : true,
//	       "auto_keep" : false
//	     },
//	     "raw" : {
//	       "value" : 0,
//	       "string" : "0"
//	     }
//	   }, {
//	     "id" : 187,
//	     "name" : "Uncorrectable_Error_Cnt",
//	     "value" : 100,
//	     "worst" : 100,
//	     "thresh" : 0,
//	     "when_failed" : "",
//	     "flags" : {
//	       "value" : 50,
//	       "string" : "-O--CK ",
//	       "prefailure" : false,
//	       "updated_online" : true,
//	       "performance" : false,
//	       "error_rate" : false,
//	       "event_count" : true,
//	       "auto_keep" : true
//	     },
//	     "raw" : {
//	       "value" : 0,
//	       "string" : "0"
//	     }
//	   }, {
//	     "id" : 190,
//	     "name" : "Airflow_Temperature_Cel",
//	     "value" : 74,
//	     "worst" : 47,
//	     "thresh" : 0,
//	     "when_failed" : "",
//	     "flags" : {
//	       "value" : 50,
//	       "string" : "-O--CK ",
//	       "prefailure" : false,
//	       "updated_online" : true,
//	       "performance" : false,
//	       "error_rate" : false,
//	       "event_count" : true,
//	       "auto_keep" : true
//	     },
//	     "raw" : {
//	       "value" : 26,
//	       "string" : "26"
//	     }
//	   }, {
//	     "id" : 195,
//	     "name" : "ECC_Error_Rate",
//	     "value" : 200,
//	     "worst" : 200,
//	     "thresh" : 0,
//	     "when_failed" : "",
//	     "flags" : {
//	       "value" : 26,
//	       "string" : "-O-RC- ",
//	       "prefailure" : false,
//	       "updated_online" : true,
//	       "performance" : false,
//	       "error_rate" : true,
//	       "event_count" : true,
//	       "auto_keep" : false
//	     },
//	     "raw" : {
//	       "value" : 0,
//	       "string" : "0"
//	     }
//	   }, {
//	     "id" : 199,
//	     "name" : "CRC_Error_Count",
//	     "value" : 99,
//	     "worst" : 99,
//	     "thresh" : 0,
//	     "when_failed" : "",
//	     "flags" : {
//	       "value" : 62,
//	       "string" : "-OSRCK ",
//	       "prefailure" : false,
//	       "updated_online" : true,
//	       "performance" : true,
//	       "error_rate" : true,
//	       "event_count" : true,
//	       "auto_keep" : true
//	     },
//	     "raw" : {
//	       "value" : 19,
//	       "string" : "19"
//	     }
//	   }, {
//	     "id" : 235,
//	     "name" : "POR_Recovery_Count",
//	     "value" : 99,
//	     "worst" : 99,
//	     "thresh" : 0,
//	     "when_failed" : "",
//	     "flags" : {
//	       "value" : 18,
//	       "string" : "-O--C- ",
//	       "prefailure" : false,
//	       "updated_online" : true,
//	       "performance" : false,
//	       "error_rate" : false,
//	       "event_count" : true,
//	       "auto_keep" : false
//	     },
//	     "raw" : {
//	       "value" : 69,
//	       "string" : "69"
//	     }
//	   }, {
//	     "id" : 241,
//	     "name" : "Total_LBAs_Written",
//	     "value" : 99,
//	     "worst" : 99,
//	     "thresh" : 0,
//	     "when_failed" : "",
//	     "flags" : {
//	       "value" : 50,
//	       "string" : "-O--CK ",
//	       "prefailure" : false,
//	       "updated_online" : true,
//	       "performance" : false,
//	       "error_rate" : false,
//	       "event_count" : true,
//	       "auto_keep" : true
//	     },
//	     "raw" : {
//	       "value" : 493358643370,
//	       "string" : "493358643370"
//	     }
//	   } ]
//	 },
//	 "spare_available" : {
//	   "current_percent" : 100,
//	   "threshold_percent" : 10
//	 },
//	 "power_on_time" : {
//	   "hours" : 36682
//	 },
//	 "power_cycle_count" : 354,
//	 "endurance_used" : {
//	   "current_percent" : 70
//	 },
//	 "temperature" : {
//	   "current" : 26
//	 }
//	}
//
// Example smartctl JSON output structure for SSD:
//
//		{
//	 "json_format_version": [
//	   1,
//	   0
//	 ],
//	 "smartctl": {
//	   "version": [
//	     7,
//	     5
//	   ],
//	   "pre_release": false,
//	   "svn_revision": "5714",
//	   "platform_info": "x86_64-linux-6.17.8-arch1-1",
//	   "build_info": "(local build)",
//	   "argv": [
//	     "smartctl",
//	     "-json",
//	     "-A",
//	     "/dev/nvme0n1"
//	   ],
//	   "exit_status": 0
//	 },
//	 "local_time": {
//	   "time_t": 1767315439,
//	   "asctime": "Fri Jan  2 01:57:19 2026 CET"
//	 },
//	 "device": {
//	   "name": "/dev/nvme0n1",
//	   "info_name": "/dev/nvme0n1",
//	   "type": "nvme",
//	   "protocol": "NVMe"
//	 },
//	 "nvme_smart_health_information_log": {
//	   "nsid": 1,
//	   "critical_warning": 0,
//	   "temperature": 55,
//	   "available_spare": 100,
//	   "available_spare_threshold": 10,
//	   "percentage_used": 2,
//	   "data_units_read": 21325779,
//	   "data_units_written": 129205897,
//	   "host_reads": 341935150,
//	   "host_writes": 2571625700,
//	   "controller_busy_time": 7846,
//	   "power_cycles": 683,
//	   "power_on_hours": 12968,
//	   "unsafe_shutdowns": 97,
//	   "media_errors": 0,
//	   "num_err_log_entries": 1860,
//	   "warning_temp_time": 0,
//	   "critical_comp_time": 0,
//	   "temperature_sensors": [
//	     55,
//	     61
//	   ]
//	 },
//	 "temperature": {
//	   "current": 55
//	 },
//	 "spare_available": {
//	   "current_percent": 100,
//	   "threshold_percent": 10
//	 },
//	 "endurance_used": {
//	   "current_percent": 2
//	 },
//	 "power_cycle_count": 683,
//	 "power_on_time": {
//	   "hours": 12968
//	 }
//	}
type SmartCtlData struct {
	JSONFormatVersion []int `json:"json_format_version"`
	Smartctl          struct {
		Version              []int    `json:"version"`
		PreRelease           bool     `json:"pre_release"`
		SvnRevision          string   `json:"svn_revision"`
		PlatformInfo         string   `json:"platform_info"`
		BuildInfo            string   `json:"build_info"`
		Argv                 []string `json:"argv"`
		DriveDatabaseVersion struct {
			String string `json:"string"`
		} `json:"drive_database_version"`
		ExitStatus int `json:"exit_status"`
	} `json:"smartctl"`
	LocalTime struct {
		TimeT   int    `json:"time_t"`
		Asctime string `json:"asctime"`
	} `json:"local_time"`
	Device struct {
		Name     string `json:"name"`
		InfoName string `json:"info_name"`
		Type     string `json:"type"`
		Protocol string `json:"protocol"`
	} `json:"device"`
	AtaSmartAttributes struct {
		Revision int `json:"revision"`
		Table    []struct {
			ID         int    `json:"id"`
			Name       string `json:"name"`
			Value      int    `json:"value"`
			Worst      int    `json:"worst"`
			Thresh     int    `json:"thresh"`
			WhenFailed string `json:"when_failed"`
			Flags      struct {
				Value         int    `json:"value"`
				String        string `json:"string"`
				Prefailure    bool   `json:"prefailure"`
				UpdatedOnline bool   `json:"updated_online"`
				Performance   bool   `json:"performance"`
				ErrorRate     bool   `json:"error_rate"`
				EventCount    bool   `json:"event_count"`
				AutoKeep      bool   `json:"auto_keep"`
			} `json:"flags"`
			Raw struct {
				Value  int    `json:"value"`
				String string `json:"string"`
			} `json:"raw"`
		} `json:"table"`
	} `json:"ata_smart_attributes"`
	NvmeSmartHealthInformationLog struct {
		Nsid                    int   `json:"nsid"`
		CriticalWarning         int   `json:"critical_warning"`
		Temperature             int   `json:"temperature"`
		AvailableSpare          int   `json:"available_spare"`
		AvailableSpareThreshold int   `json:"available_spare_threshold"`
		PercentageUsed          int   `json:"percentage_used"`
		DataUnitsRead           int   `json:"data_units_read"`
		DataUnitsWritten        int   `json:"data_units_written"`
		HostReads               int   `json:"host_reads"`
		HostWrites              int64 `json:"host_writes"`
		ControllerBusyTime      int   `json:"controller_busy_time"`
		PowerCycles             int   `json:"power_cycles"`
		PowerOnHours            int   `json:"power_on_hours"`
		UnsafeShutdowns         int   `json:"unsafe_shutdowns"`
		MediaErrors             int   `json:"media_errors"`
		NumErrLogEntries        int   `json:"num_err_log_entries"`
		WarningTempTime         int   `json:"warning_temp_time"`
		CriticalCompTime        int   `json:"critical_comp_time"`
		TemperatureSensors      []int `json:"temperature_sensors"`
	} `json:"nvme_smart_health_information_log"`
	SpareAvailable struct {
		CurrentPercent   int `json:"current_percent"`
		ThresholdPercent int `json:"threshold_percent"`
	} `json:"spare_available"`
	PowerOnTime struct {
		Hours int `json:"hours"`
	} `json:"power_on_time"`
	PowerCycleCount int `json:"power_cycle_count"`
	EnduranceUsed   struct {
		CurrentPercent int `json:"current_percent"`
	} `json:"endurance_used"`
	Temperature struct {
		Current int `json:"current"`
	} `json:"temperature"`
}

// sudo nvme smart-log /dev/nvme0n1 -o json -H
//
// Example NVMe smart log structure
//
//	{
//	 "critical_warning":{
//	   "value":0,
//	   "available_spare":0,
//	   "temp_threshold":0,
//	   "reliability_degraded":0,
//	   "ro":0,
//	   "vmbu_failed":0,
//	   "pmr_ro":0
//	 },
//	 "temperature":328,
//	 "avail_spare":100,
//	 "spare_thresh":10,
//	 "percent_used":2,
//	 "endurance_grp_critical_warning_summary":0,
//	 "data_units_read":21331123,
//	 "data_units_written":129415397,
//	 "host_read_commands":342132364,
//	 "host_write_commands":2576184016,
//	 "controller_busy_time":7856,
//	 "power_cycles":683,
//	 "power_on_hours":12993,
//	 "unsafe_shutdowns":97,
//	 "media_errors":0,
//	 "num_err_log_entries":1860,
//	 "warning_temp_time":0,
//	 "critical_comp_time":0,
//	 "temperature_sensor_1":328,
//	 "temperature_sensor_2":334,
//	 "thm_temp1_trans_count":0,
//	 "thm_temp2_trans_count":0,
//	 "thm_temp1_total_time":0,
//	 "thm_temp2_total_time":0
//	}
type NvmeData struct {
	CriticalWarning struct {
		Value               int `json:"value"`
		AvailableSpare      int `json:"available_spare"`
		TempThreshold       int `json:"temp_threshold"`
		ReliabilityDegraded int `json:"reliability_degraded"`
		Ro                  int `json:"ro"`
		VmbuFailed          int `json:"vmbu_failed"`
		PmrRo               int `json:"pmr_ro"`
	} `json:"critical_warning"`
	Temperature                        int `json:"temperature"`
	AvailSpare                         int `json:"avail_spare"`
	SpareThresh                        int `json:"spare_thresh"`
	PercentUsed                        int `json:"percent_used"`
	EnduranceGrpCriticalWarningSummary int `json:"endurance_grp_critical_warning_summary"`
	DataUnitsRead                      int `json:"data_units_read"`
	DataUnitsWritten                   int `json:"data_units_written"`
	HostReadCommands                   int `json:"host_read_commands"`
	HostWriteCommands                  int `json:"host_write_commands"`
	ControllerBusyTime                 int `json:"controller_busy_time"`
	PowerCycles                        int `json:"power_cycles"`
	PowerOnHours                       int `json:"power_on_hours"`
	UnsafeShutdowns                    int `json:"unsafe_shutdowns"`
	MediaErrors                        int `json:"media_errors"`
	NumErrLogEntries                   int `json:"num_err_log_entries"`
	WarningTempTime                    int `json:"warning_temp_time"`
	CriticalCompTime                   int `json:"critical_comp_time"`
	TemperatureSensor1                 int `json:"temperature_sensor_1"`
	TemperatureSensor2                 int `json:"temperature_sensor_2"`
	ThmTemp1TransCount                 int `json:"thm_temp1_trans_count"`
	ThmTemp2TransCount                 int `json:"thm_temp2_trans_count"`
	ThmTemp1TotalTime                  int `json:"thm_temp1_total_time"`
	ThmTemp2TotalTime                  int `json:"thm_temp2_total_time"`
}

type DiskInfo struct {
	/** f.ex. ata-ST4000DM004-2CV104_Z301XXXX */
	Name string
	/** f.ex. /dev/sdb */
	Path string
}

func GetDisks() ([]DiskInfo, error) {
	result := []DiskInfo{}

	files, err := os.ReadDir(DiskByIdPath)
	if err != nil {
		return result, err
	}

	regex := regexp.MustCompile(`^(ata|^nvme|^scsi|^wwn)-.*$`)
	suffixNoPartRegex := regexp.MustCompile(`-part[0-9]+$`)

	for _, f := range files {
		if f.Type() != os.ModeSymlink {
			continue
		}

		if !regex.MatchString(f.Name()) {
			continue
		}

		if suffixNoPartRegex.MatchString(f.Name()) {
			// ignore partition entries
			continue
		}

		// resolve symlink
		entryPath := DiskByIdPath + f.Name()
		linkTarget, err := os.Readlink(entryPath)
		if err != nil {
			continue
		}
		// resolve relative paths (../../sdb -> /dev/sdb)
		linkTarget = filepath.Join(DiskByIdPath, linkTarget)

		// if link target is already in the result list, skip it
		skip := false
		for _, d := range result {
			if d.Path == linkTarget {
				skip = true
				break
			}
		}
		if skip {
			continue
		}

		disk := DiskInfo{
			Name: f.Name(),
			Path: linkTarget,
		}
		result = append(result, disk)
	}

	// sort by Name
	sort.SliceStable(result, func(i, j int) bool {
		a := result[i]
		b := result[j]

		result := 0
		result = strings.Compare(strings.ToLower(a.Name), strings.ToLower(b.Name))

		if result <= 0 {
			return true
		} else {
			return false
		}
	})

	return result, nil
}

func (d DiskInfo) GetSmartCtlData() (SmartCtlData, error) {
	// example command:
	// smartctl -json -A /dev/sdg -v 1,raw48:54 -v 7,raw48:54 -v 241,raw48:54 -v 242,raw48:54

	var args = []string{
		"smartctl",
		"-json",
		"-A",
		d.Path,
	}

	var vendorSpecificAttrs = []string{}
	if strings.HasPrefix(d.Name, "ata-ST") {
		vendorSpecificAttrs = []string{
			"-v", "1,raw48:54",
			"-v", "7,raw48:54",
			"-v", "241,raw48:54",
			"-v", "242,raw48:54",
		}
	}
	args = append(args, vendorSpecificAttrs...)

	output, err := ExecCommand("sudo", args...)
	if err != nil {
		return SmartCtlData{}, err
	}

	var result SmartCtlData
	err = json.Unmarshal([]byte(output), &result)
	return result, err
}

func (d DiskInfo) GetNvmeSmartLog() (NvmeData, error) {
	// example command:
	// sudo nvme smart-log /dev/nvme0n1 -o json -H

	var args = []string{
		"nvme",
		"smart-log",
		d.Path,
		"-o",
		"json",
		"-H",
	}

	output, err := ExecCommand("sudo", args...)
	if err != nil {
		return NvmeData{}, err
	}

	var result NvmeData
	err = json.Unmarshal([]byte(output), &result)
	return result, err
}
