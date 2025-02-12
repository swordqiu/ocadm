// Copyright 2019 Yunion
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package modules

import "yunion.io/x/onecloud/pkg/mcclient/modulebase"

var (
	BillAnalysises         modulebase.ResourceManager
	BillsUpgradeAnalysises modulebase.ResourceManager
)

func init() {
	BillAnalysises = NewMeterManager("bill_analysis", "bill_analysises",
		[]string{"stat_date", "stat_value", "res_name", "res_type", "project_name", "res_fee"},
		[]string{},
	)

	BillsUpgradeAnalysises = NewMeterManager("billsanalysis", "billsanalysises",
		[]string{"project", "project_id", "domain", "domain_id", "amount", "year_amount"},
		[]string{})
	register(&BillAnalysises)
	register(&BillsUpgradeAnalysises)
}
