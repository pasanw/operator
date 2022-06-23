// Copyright (c) 2020-2022 Tigera, Inc. All rights reserved.

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

// Components defined here are required to be kept in sync with
// config/enterprise_versions.yml

package components

var (
	EnterpriseRelease string = "development"

	ComponentAPIServer = component{
		Version: "development",
		Image:   "tigera/cnx-apiserver",
	}

	ComponentComplianceBenchmarker = component{
		Version: "development",
		Image:   "tigera/compliance-benchmarker",
	}

	ComponentComplianceController = component{
		Version: "development",
		Image:   "tigera/compliance-controller",
	}

	ComponentComplianceReporter = component{
		Version: "development",
		Image:   "tigera/compliance-reporter",
	}

	ComponentComplianceServer = component{
		Version: "development",
		Image:   "tigera/compliance-server",
	}

	ComponentComplianceSnapshotter = component{
		Version: "development",
		Image:   "tigera/compliance-snapshotter",
	}

	ComponentDeepPacketInspection = component{
		Version: "development",
		Image:   "tigera/deep-packet-inspection",
	}

	ComponentEckElasticsearch = component{
		Version: "7.16.2",
		Image:   "tigera/elasticsearch",
	}

	ComponentEckKibana = component{
		Version: "7.16.2",
		Image:   "tigera/kibana",
	}

	ComponentElasticTseeInstaller = component{
		Version: "development",
		Image:   "tigera/intrusion-detection-job-installer",
	}

	ComponentElasticsearch = component{
		Version: "development",
		Image:   "tigera/elasticsearch",
	}

	ComponentECKElasticsearchOperator = component{
		Version: "1.8.0",
		Image:   "tigera/eck-operator",
	}

	ComponentElasticsearchOperator = component{
		Version: "development",
		Image:   "tigera/eck-operator",
	}

	ComponentEsCurator = component{
		Version: "development",
		Image:   "tigera/es-curator",
	}

	ComponentEsProxy = component{
		Version: "development",
		Image:   "tigera/es-proxy",
	}

	ComponentESGateway = component{
		Version: "development",
		Image:   "tigera/es-gateway",
	}

	ComponentFluentd = component{
		Version: "development",
		Image:   "tigera/fluentd",
	}

	ComponentFluentdWindows = component{
		Version: "development",
		Image:   "tigera/fluentd-windows",
	}

	ComponentGuardian = component{
		Version: "development",
		Image:   "tigera/guardian",
	}

	ComponentIntrusionDetectionController = component{
		Version: "development",
		Image:   "tigera/intrusion-detection-controller",
	}

	ComponentAnomalyDetectionJobs = component{
		Version: "development",
		Image:   "tigera/anomaly_detection_jobs",
	}

	ComponentAnomalyDetectionAPI = component{
		Version: "development",
		Image:   "tigera/anomaly-detection-api",
	}

	ComponentKibana = component{
		Version: "development",
		Image:   "tigera/kibana",
	}

	ComponentManager = component{
		Version: "development",
		Image:   "tigera/cnx-manager",
	}

	ComponentDex = component{
		Version: "development",
		Image:   "tigera/dex",
	}

	ComponentManagerProxy = component{
		Version: "development",
		Image:   "tigera/voltron",
	}

	ComponentPacketCapture = component{
		Version: "development",
		Image:   "tigera/packetcapture",
	}

	ComponentL7Collector = component{
		Version: "development",
		Image:   "tigera/l7-collector",
	}

	ComponentEnvoyProxy = component{
		Version: "development",
		Image:   "tigera/envoy",
	}

	ComponentDikastes = component{
		Version: "development",
		Image:   "tigera/dikastes",
	}

	ComponentCoreOSPrometheus = component{
		Version: "v2.32.0",
		Image:   "tigera/prometheus",
	}

	ComponentPrometheus = component{
		Version: "development",
		Image:   "tigera/prometheus",
	}

	ComponentTigeraPrometheusService = component{
		Version: "development",
		Image:   "tigera/prometheus-service",
	}

	ComponentCoreOSAlertmanager = component{
		Version: "v0.23.0",
		Image:   "tigera/alertmanager",
	}

	ComponentPrometheusAlertmanager = component{
		Version: "development",
		Image:   "tigera/alertmanager",
	}

	ComponentQueryServer = component{
		Version: "development",
		Image:   "tigera/cnx-queryserver",
	}

	ComponentTigeraKubeControllers = component{
		Version: "development",
		Image:   "tigera/kube-controllers",
	}

	ComponentTigeraNode = component{
		Version: "development",
		Image:   "tigera/cnx-node",
	}

	ComponentTigeraTypha = component{
		Version: "development",
		Image:   "tigera/typha",
	}

	ComponentTigeraCNI = component{
		Version: "development",
		Image:   "tigera/cni",
	}

	ComponentCloudControllers = component{
		Version: "development",
		Image:   "tigera/cloud-controllers",
	}

	ComponentElasticsearchMetrics = component{
		Version: "development",
		Image:   "tigera/elasticsearch-metrics",
	}

	ComponentTigeraWindows = component{
		Version: "development",
		Image:   "tigera/calico-windows-upgrade",
	}
	EnterpriseComponents = []component{
		ComponentAPIServer,
		ComponentComplianceBenchmarker,
		ComponentComplianceController,
		ComponentComplianceReporter,
		ComponentComplianceServer,
		ComponentComplianceSnapshotter,
		ComponentDeepPacketInspection,
		ComponentEckElasticsearch,
		ComponentEckKibana,
		ComponentElasticTseeInstaller,
		ComponentElasticsearch,
		ComponentECKElasticsearchOperator,
		ComponentElasticsearchOperator,
		ComponentEsCurator,
		ComponentEsProxy,
		ComponentFluentd,
		ComponentFluentdWindows,
		ComponentGuardian,
		ComponentIntrusionDetectionController,
		ComponentAnomalyDetectionJobs,
		ComponentAnomalyDetectionAPI,
		ComponentKibana,
		ComponentManager,
		ComponentDex,
		ComponentManagerProxy,
		ComponentPacketCapture,
		ComponentL7Collector,
		ComponentEnvoyProxy,
		ComponentCoreOSPrometheus,
		ComponentPrometheus,
		ComponentTigeraPrometheusService,
		ComponentCoreOSAlertmanager,
		ComponentPrometheusAlertmanager,
		ComponentQueryServer,
		ComponentTigeraKubeControllers,
		ComponentTigeraNode,
		ComponentTigeraTypha,
		ComponentTigeraCNI,
		ComponentCloudControllers,
		ComponentElasticsearchMetrics,
		ComponentESGateway,
		ComponentTigeraWindows,
		ComponentDikastes,
	}
)
