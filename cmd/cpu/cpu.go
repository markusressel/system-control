/*
 * system-control
 * Copyright (c) 2019. Markus Ressel
 *
 * This program is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Affero General Public License as
 * published by the Free Software Foundation, either version 3 of the
 * License, or (at your option) any later version.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU Affero General Public License for more details.
 *
 * You should have received a copy of the GNU Affero General Public License
 * along with this program.  If not, see <https://www.gnu.org/licenses/>.
 */
package cpu

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/elliotchance/orderedmap/v2"
	"github.com/markusressel/system-control/internal/util"
	"github.com/spf13/cobra"
)

var Command = &cobra.Command{
	Use:   "cpu",
	Short: "Control CPU settings",
	Long:  ``,
	RunE: func(cmd *cobra.Command, args []string) error {
		cpuInfoList, err := GetCpuInfo()
		if err != nil {
			return err
		}

		for i, cpuInfo := range cpuInfoList {
			properties := orderedmap.NewOrderedMap[string, string]()
			properties.Set("Vendor ID", cpuInfo.VendorId)
			properties.Set("Model Name", cpuInfo.ModelName)

			util.PrintFormattedTableOrdered(fmt.Sprintf("%d", cpuInfo.Index), properties)

			if i < len(cpuInfoList)-1 {
				fmt.Println()
			}
		}

		return nil
	},
}

// CpuInfo
// processor       : 0
// vendor_id       : GenuineIntel
// cpu family      : 6
// model           : 141
// model name      : 11th Gen Intel(R) Core(TM) i9-11900H @ 2.50GHz
// stepping        : 1
// microcode       : 0x52
// cpu MHz         : 2036.474
// cache size      : 24576 KB
// physical id     : 0
// siblings        : 16
// core id         : 0
// cpu cores       : 8
// apicid          : 0
// initial apicid  : 0
// fpu             : yes
// fpu_exception   : yes
// cpuid level     : 27
// wp              : yes
// flags           : fpu vme de pse tsc msr pae mce cx8 apic sep mtrr pge mca cmov pat pse36 clflush dts acpi mmx fxsr sse sse2 ss ht tm pbe syscall nx pdpe1gb rdtscp lm constant_tsc art arch_perfmon pebs bts rep_good nopl xtopology nonstop_tsc cpuid aperfmpe
// rf tsc_known_freq pni pclmulqdq dtes64 monitor ds_cpl vmx est tm2 ssse3 sdbg fma cx16 xtpr pdcm pcid sse4_1 sse4_2 x2apic movbe popcnt tsc_deadline_timer aes xsave avx f16c rdrand lahf_lm abm 3dnowprefetch cpuid_fault epb cat_l2 cdp_l2 ssbd ibrs ibpb stibp
// ibrs_enhanced tpr_shadow flexpriority ept vpid ept_ad fsgsbase tsc_adjust bmi1 avx2 smep bmi2 erms invpcid rdt_a avx512f avx512dq rdseed adx smap avx512ifma clflushopt clwb intel_pt avx512cd sha_ni avx512bw avx512vl xsaveopt xsavec xgetbv1 xsaves split_lo
// ck_detect user_shstk dtherm ida arat pln pts hwp hwp_notify hwp_act_window hwp_epp hwp_pkg_req vnmi avx512vbmi umip pku ospke avx512_vbmi2 gfni vaes vpclmulqdq avx512_vnni avx512_bitalg avx512_vpopcntdq rdpid movdiri movdir64b fsrm avx512_vp2intersect md_c
// lear ibt flush_l1d arch_capabilities
// vmx flags       : vnmi preemption_timer posted_intr invvpid ept_x_only ept_ad ept_1gb flexpriority apicv tsc_offset vtpr mtf vapic ept vpid unrestricted_guest vapic_reg vid ple pml ept_violation_ve ept_mode_based_exec tsc_scaling
// bugs            : spectre_v1 spectre_v2 spec_store_bypass swapgs eibrs_pbrsb gds bhi
// bogomips        : 4993.00
// clflush size    : 64
// cache_alignment : 64
// address sizes   : 39 bits physical, 48 bits virtual
// power management:
type CpuInfo struct {
	Index     int
	VendorId  string
	ModelName string
}

func GetCpuInfo() ([]CpuInfo, error) {
	result, err := util.ExecCommand(
		"cat",
		"/proc/cpuinfo",
	)
	if err != nil {
		return nil, err
	}

	// split on empty lines
	entries := strings.Split(result, "\n\n")

	cpuInfoList := make([]CpuInfo, 0)
	for _, entry := range entries {
		cpuInfo, err := parseCpuInfoEntry(entry)
		if err != nil {
			continue
		}
		cpuInfoList = append(cpuInfoList, cpuInfo)
	}

	return cpuInfoList, nil
}

func parseCpuInfoEntry(entry string) (CpuInfo, error) {
	lines := strings.Split(entry, "\n")

	cpuInfoMap := make(map[string]string)
	for _, line := range lines {
		if strings.Contains(line, ":") {
			parts := strings.SplitN(line, ":", 2)
			key := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])
			cpuInfoMap[key] = value
		}
	}

	index, _ := strconv.Atoi(cpuInfoMap["processor"])
	cpuInfo := CpuInfo{
		Index:     index,
		VendorId:  cpuInfoMap["vendor_id"],
		ModelName: cpuInfoMap["model name"],
	}

	return cpuInfo, nil
}
