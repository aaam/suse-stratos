package kubernetes

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/url"
	"strconv"

	"errors"

	"github.com/cloudfoundry-incubator/stratos/src/jetstream/repository/interfaces"
	"github.com/labstack/echo"
	log "github.com/sirupsen/logrus"

	"github.com/cloudfoundry-incubator/stratos/src/jetstream/plugins/kubernetes/auth"
)

// KubernetesSpecification is the endpoint that adds Kubernetes support to the backend
type KubernetesSpecification struct {
	portalProxy  interfaces.PortalProxy
	endpointType string
}

type KubeStatus struct {
	Kind       string      `json:"kind"`
	ApiVersion string      `json:"apiVersion"`
	Metadata   interface{} `json:"metadata"`
	Status     string      `json:"status"`
	Message    string      `json:"message"`
	Reason     string      `json:"reason"`
	Details    interface{} `json:"details"`
	Code       int         `json:"code"`
}

type KubeAPIVersions struct {
	Kind                       string        `json:"kind"`
	Versions                   []string      `json:"versions"`
	ServerAddressByClientCIDRs []interface{} `json:"serverAddressByClientCIDRs"`
}

const (
	kubeEndpointType    = "k8s"
	defaultKubeClientID = "K8S_CLIENT"

	// kubeDashboardPluginConfigSetting is config value send back to the client to indicate if the kube dashboard can be navigated to
	kubeDashboardPluginConfigSetting = "kubeDashboardEnabled"
)

func Init(portalProxy interfaces.PortalProxy) (interfaces.StratosPlugin, error) {
	return &KubernetesSpecification{portalProxy: portalProxy, endpointType: kubeEndpointType}, nil
}

func (c *KubernetesSpecification) GetEndpointPlugin() (interfaces.EndpointPlugin, error) {
	return c, nil
}

func (c *KubernetesSpecification) GetRoutePlugin() (interfaces.RoutePlugin, error) {
	return c, nil
}

func (c *KubernetesSpecification) GetMiddlewarePlugin() (interfaces.MiddlewarePlugin, error) {
	return nil, errors.New("Not implemented!")
}

func (c *KubernetesSpecification) GetType() string {
	return kubeEndpointType
}

func (c *KubernetesSpecification) GetClientId() string {
	return c.portalProxy.Env().String(defaultKubeClientID, "k8s")
}

func (c *KubernetesSpecification) Register(echoContext echo.Context) error {
	log.Debug("Kubernetes Register...")
	return c.portalProxy.RegisterEndpoint(echoContext, c.Info)
}

func (c *KubernetesSpecification) Validate(userGUID string, cnsiRecord interfaces.CNSIRecord, tokenRecord interfaces.TokenRecord) error {
	log.Debugf("Validating Kubernetes endpoint connection for user: %s", userGUID)
	response, err := c.portalProxy.DoProxySingleRequest(cnsiRecord.GUID, userGUID, "GET", "api/v1/pods?limit=1", nil, nil)
	if err != nil {
		return err
	}

	if response.StatusCode >= 400 {
		return fmt.Errorf("Unable to connect to endpoint: %s", response.Error.Error())
	}

	return nil
}

func (c *KubernetesSpecification) Connect(ec echo.Context, cnsiRecord interfaces.CNSIRecord, userID string) (*interfaces.TokenRecord, bool, error) {
	log.Debug("Kubernetes Connect...")

	connectType := ec.FormValue("connect_type")

	var authProvider = c.FindAuthProvider(connectType)
	if authProvider == nil {
		return nil, false, errors.New("Unsupported Auth connection type for Kubernetes endpoint")
	}

	tokenRecord, _, err := authProvider.FetchToken(cnsiRecord, ec)
	if err != nil {
		return nil, false, err
	}

	return tokenRecord, false, nil
}

// Init the Kubernetes Jetstream plugin
func (c *KubernetesSpecification) Init() error {

	c.AddAuthProvider(auth.InitGKEKubeAuth(c.portalProxy))
	c.AddAuthProvider(auth.InitAWSKubeAuth(c.portalProxy))
	c.AddAuthProvider(auth.InitCertKubeAuth(c.portalProxy))
	c.AddAuthProvider(auth.InitAzureKubeAuth(c.portalProxy))
	c.AddAuthProvider(auth.InitOIDCKubeAuth(c.portalProxy))
	c.AddAuthProvider(auth.InitKubeConfigAuth(c.portalProxy))
	c.AddAuthProvider(auth.InitKubeTokenAuth(c.portalProxy))

	// Kube dashboard is enabled by Tech Preview mode
	c.portalProxy.GetConfig().PluginConfig[kubeDashboardPluginConfigSetting] = strconv.FormatBool(c.portalProxy.GetConfig().EnableTechPreview)

	return nil
}

func (c *KubernetesSpecification) AddAdminGroupRoutes(echoGroup *echo.Group) {
	// no-op
}

func (c *KubernetesSpecification) AddSessionGroupRoutes(echoGroup *echo.Group) {

	// Kubernetes Dashboard Proxy
	echoGroup.Any("/apps/kubedash/ui/:guid/*", c.kubeDashboardProxy)

	echoGroup.GET("/kubedash/:guid/login", c.kubeDashboardLogin)
	echoGroup.GET("/kubedash/:guid/status", c.kubeDashboardStatus)

	echoGroup.POST("/kubedash/:guid/serviceAccount", c.kubeDashboardCreateServiceAccount)
	echoGroup.DELETE("/kubedash/:guid/serviceAccount", c.kubeDashboardDeleteServiceAccount)

	echoGroup.POST("/kubedash/:guid/installation", c.kubeDashboardInstallDashboard)
	echoGroup.DELETE("/kubedash/:guid/installation", c.kubeDashboardDeleteDashboard)

	// Helm Routes
	echoGroup.GET("/helm/releases", c.ListReleases)
	echoGroup.POST("/helm/install", c.InstallRelease)
	echoGroup.DELETE("/helm/releases/:endpoint/:namespace/:name", c.DeleteRelease)
	echoGroup.GET("/helm/releases/:endpoint/:namespace/:name/status", c.GetReleaseStatus)
	echoGroup.GET("/helm/releases/:endpoint/:namespace/:name", c.GetRelease)

}

func (c *KubernetesSpecification) Info(apiEndpoint string, skipSSLValidation bool) (interfaces.CNSIRecord, interface{}, error) {

	log.Debug("Kubernetes Info")
	var v2InfoResponse interfaces.V2Info
	var newCNSI interfaces.CNSIRecord

	newCNSI.CNSIType = kubeEndpointType

	_, err := url.Parse(apiEndpoint)
	if err != nil {
		return newCNSI, nil, err
	}

	log.Debug("Request Kube API Versions")
	var httpClient = c.portalProxy.GetHttpClient(skipSSLValidation)
	res, err := httpClient.Get(apiEndpoint + "/api")
	if err != nil {
		// This should ultimately catch 503 cert errors
		return newCNSI, nil, err
	}

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return newCNSI, nil, err
	}

	if res.StatusCode < 400 {
		// No auth on kube set up, expect a successful APIVersions response - KubeAPIVersions
		log.Debug("Kube API Versions Succeeded")
		apiVersions := KubeAPIVersions{}
		err := json.Unmarshal(body, &apiVersions)
		if err != nil {
			return newCNSI, nil, fmt.Errorf("Failed to parse output as kube kind APIVersions: %+v", err)
		}
		if apiVersions.Kind != "APIVersions" {
			return newCNSI, nil, fmt.Errorf("Failed to parse output as kube kind APIVersions: %+v", apiVersions)
		}
	} else if res.StatusCode == 403 {
		// Expect an auth failed response - KubeStatus
		log.Debug("Kube API Versions Failed (403)")
		kubeStatus := KubeStatus{}
		err := json.Unmarshal(body, &kubeStatus)
		if err != nil {
			return newCNSI, nil, fmt.Errorf("Failed to parse 403 output as kube kind status: %+v", err)
		}
		if kubeStatus.Kind != "Status" {
			return newCNSI, nil, fmt.Errorf("Failed to parse 403 output as kube kind status: %+v", kubeStatus)
		}
	} else {
		return newCNSI, nil, fmt.Errorf("Dissallowed response code from `/api` call: %+v", res.StatusCode)
	}

	log.Debug("Kube API Versions Acceptable Response")
	newCNSI.TokenEndpoint = apiEndpoint
	newCNSI.AuthorizationEndpoint = apiEndpoint

	return newCNSI, v2InfoResponse, nil
}

func (c *KubernetesSpecification) UpdateMetadata(info *interfaces.Info, userGUID string, echoContext echo.Context) {
}
