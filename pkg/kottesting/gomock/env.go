package kotmocks

import (
	"io"

	"github.com/golang/mock/gomock"
	apiruntime "k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
)

type MockedEnv struct {
	Scheme  *apiruntime.Scheme
	Client  *MockClient
	Manager *MockManager
}

func NewEnv(mCtrl *gomock.Controller, logOutput io.Writer) MockedEnv {
	scheme := apiruntime.NewScheme()
	if err := clientgoscheme.AddToScheme(scheme); err != nil {
		panic(err)
	}

	client := NewMockClient(mCtrl)

	mgr := NewMockManager(mCtrl)
	mgr.EXPECT().GetScheme().Return(scheme).AnyTimes()
	mgr.EXPECT().GetClient().Return(client).AnyTimes()

	logger := zap.New(zap.WriteTo(logOutput), zap.UseDevMode(true))
	mgr.EXPECT().GetLogger().Return(logger).AnyTimes()

	return MockedEnv{scheme, client, mgr}
}
