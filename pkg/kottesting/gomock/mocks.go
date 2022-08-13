package kotmocks

//go:generate mockgen -package kotmocks -destination ./client.go -source ../../kotclient/client.go Client
//go:generate mockgen -package kotmocks -destination ./manager.go sigs.k8s.io/controller-runtime Manager
