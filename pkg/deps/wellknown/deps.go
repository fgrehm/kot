package wellknown

import (
	"github.com/fgrehm/kot/pkg/deps"
	"github.com/fgrehm/kot/pkg/kotclient"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
)

const (
	mgrKey    = "kot-ctrl-runtime-mgr"
	clientKey = "kot-client"
	schemeKey = "kot-scheme"
)

func SetManager(mgr ctrl.Manager) {
	deps.Set(mgrKey, mgr)
	deps.Set(schemeKey, mgr.GetScheme())
	SetClient(kotclient.Decorate(mgr.GetClient()))
}

func Manager(ctn interface{}) ctrl.Manager {
	return deps.Get(ctn, mgrKey).(ctrl.Manager)
}

func SetScheme(scheme *runtime.Scheme) {
	deps.Set(schemeKey, scheme)
}

func Scheme(ctn interface{}) *runtime.Scheme {
	return deps.Get(ctn, schemeKey).(*runtime.Scheme)
}

func SetClient(client kotclient.Client) {
	deps.Set(clientKey, client)
}

func Client(ctn interface{}) kotclient.Client {
	return deps.Get(ctn, clientKey).(kotclient.Client)
}
