package reconcile_test

import (
	// "errors"

	"github.com/fgrehm/kot/pkg/deps"
	"github.com/fgrehm/kot/pkg/reconcile"
	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	runtimeclient "sigs.k8s.io/controller-runtime/pkg/client"
	runtimeevent "sigs.k8s.io/controller-runtime/pkg/event"
	runtimesource "sigs.k8s.io/controller-runtime/pkg/source"
)

var _ = Describe("Watcher", func() {
	var (
		mCtrl *gomock.Controller
	)

	BeforeEach(func() {
		mCtrl = gomock.NewController(GinkgoT())
	})

	AfterEach(func() {
		mCtrl.Finish()
	})

	Describe("ResourceWatcher", func() {
		var (
			watcher *reconcile.ResourceWatcher
		)

		BeforeEach(func() {
			watcher = &reconcile.ResourceWatcher{ResourceWatcherConfig: &reconcile.ResourceWatcherConfig{}}
		})

		Describe("Source", func() {
			It("builds a source.Kind for the given type", func() {
				watcher.Watches = &corev1.Pod{}
				Expect(watcher.Source()).To(Equal(&runtimesource.Kind{Type: &corev1.Pod{}}))
			})
		})

		Describe("Handler", func() {
			It("calls the enqueue function", func() {
				called := false
				watcher.Enqueue = func(ctn deps.Container, obj runtimeclient.Object) ([]ctrl.Request, error) {
					called = true
					return []ctrl.Request{}, nil
				}

				ctn := deps.Build()
				deps.Inject(ctn, watcher)

				watcher.Handler().Create(runtimeevent.CreateEvent{}, nil)
				Expect(called).To(BeTrue())
			})
		})
	})
})
