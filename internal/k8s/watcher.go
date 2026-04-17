package k8s

import (
	"context"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
)

// Watcher maintains a local cache of running Pods on the host node via Kubernetes Informers.
type Watcher struct {
	mu           sync.RWMutex
	containerMap map[string]PodContext
	nodeName     string
}

// PodContext holds the resolved metadata for a container.
type PodContext struct {
	Namespace string
	PodName   string
}

// NewWatcher initializes a Kubernetes client and starts an informer filtering for the local node.
func NewWatcher(ctx context.Context) (*Watcher, error) {
	config, err := rest.InClusterConfig()
	if err != nil {
		// Fallback to kubeconfig
		kubeconfig := os.Getenv("KUBECONFIG")
		if kubeconfig == "" {
			home, _ := os.UserHomeDir()
			kubeconfig = home + "/.kube/config"
		}
		config, err = clientcmd.BuildConfigFromFlags("", kubeconfig)
		if err != nil {
			return nil, fmt.Errorf("failed to load kubeconfig: %w", err)
		}
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create k8s client: %w", err)
	}

	nodeName := os.Getenv("NODE_NAME")
	if nodeName == "" {
		nodeName, _ = os.Hostname() // Best effort if flag is omitted
	}

	w := &Watcher{
		containerMap: make(map[string]PodContext),
		nodeName:     nodeName,
	}

	// Optimize informer to only fetch pods running on THIS specific node.
	tweakListOptions := func(options *metav1.ListOptions) {
		if nodeName != "" {
			options.FieldSelector = fmt.Sprintf("spec.nodeName=%s", nodeName)
		}
	}

	factory := informers.NewSharedInformerFactoryWithOptions(clientset, time.Hour*24, informers.WithTweakListOptions(tweakListOptions))
	podInformer := factory.Core().V1().Pods().Informer()

	_, err = podInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			if pod, ok := obj.(*corev1.Pod); ok {
				w.updatePod(pod)
			}
		},
		UpdateFunc: func(old, new interface{}) {
			if pod, ok := new.(*corev1.Pod); ok {
				w.updatePod(pod)
			}
		},
		DeleteFunc: func(obj interface{}) {
			if pod, ok := obj.(*corev1.Pod); ok {
				w.deletePod(pod)
			}
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to add event handler: %w", err)
	}

	go podInformer.Run(ctx.Done())

	// Wait for the cache to sync before proceeding so we don't drop early events.
	if !cache.WaitForCacheSync(ctx.Done(), podInformer.HasSynced) {
		return nil, fmt.Errorf("failed to sync k8s pod cache")
	}

	return w, nil
}

func (w *Watcher) updatePod(pod *corev1.Pod) {
	w.mu.Lock()
	defer w.mu.Unlock()

	ctx := PodContext{
		Namespace: pod.Namespace,
		PodName:   pod.Name,
	}

	mapStatuses := func(statuses []corev1.ContainerStatus) {
		for _, s := range statuses {
			if s.ContainerID != "" {
				// ContainerID comes in as "docker://hex" or "containerd://hex"
				parts := strings.Split(s.ContainerID, "://")
				id := parts[len(parts)-1]
				w.containerMap[id] = ctx
			}
		}
	}
	mapStatuses(pod.Status.ContainerStatuses)
	mapStatuses(pod.Status.InitContainerStatuses)
	mapStatuses(pod.Status.EphemeralContainerStatuses)
}

func (w *Watcher) deletePod(pod *corev1.Pod) {
	w.mu.Lock()
	defer w.mu.Unlock()

	unmapStatuses := func(statuses []corev1.ContainerStatus) {
		for _, s := range statuses {
			if s.ContainerID != "" {
				parts := strings.Split(s.ContainerID, "://")
				id := parts[len(parts)-1]
				delete(w.containerMap, id)
			}
		}
	}
	unmapStatuses(pod.Status.ContainerStatuses)
	unmapStatuses(pod.Status.InitContainerStatuses)
	unmapStatuses(pod.Status.EphemeralContainerStatuses)
}

// Resolve maps a container ID to its Kubernetes Pod context.
func (w *Watcher) Resolve(containerID string) (string, string, bool) {
	if containerID == "" {
		return "", "", false
	}
	w.mu.RLock()
	defer w.mu.RUnlock()
	ctx, ok := w.containerMap[containerID]
	return ctx.Namespace, ctx.PodName, ok
}
