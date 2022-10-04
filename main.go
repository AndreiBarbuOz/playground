package main

import (
	"context"
	"fmt"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kuberuntime "k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes/fake"
	clienttesting "k8s.io/client-go/testing"
	"k8s.io/client-go/tools/cache"
	watchtools "k8s.io/client-go/tools/watch"
	"time"
)

const (
	namespace = "default"
)

func podRunningAndReady(event watch.Event) (bool, error) {
	switch event.Type {
	case watch.Deleted:
		return false, fmt.Errorf("pod already deleted")
	}
	switch t := event.Object.(type) {
	case *corev1.Pod:
		switch t.Status.Phase {
		case corev1.PodFailed, corev1.PodSucceeded:
			return false, fmt.Errorf("pod %s/%s already completed", t.Namespace, t.Name)
		case corev1.PodRunning:
			conditions := t.Status.Conditions
			if conditions == nil {
				return false, nil
			}
			for i := range conditions {
				if conditions[i].Type == corev1.PodReady &&
					conditions[i].Status == corev1.ConditionTrue {
					return true, nil
				}
			}
		}
	}
	return false, nil
}

// waitForPod watches the given pod until the exitCondition is true
func waitForPod(timeout time.Duration) (*corev1.Pod, error) {
	ctx, cancel := watchtools.ContextWithOptionalTimeout(context.Background(), timeout)
	defer cancel()

	client := fake.NewSimpleClientset()
	podClient := client.CoreV1()
	//labelSelector, err := labels.NewRequirement("app", selection.Equals, []string{"istioValidation"})
	fmt.Printf("creating the ListWatch\n")
	lw := &cache.ListWatch{
		ListFunc: func(options metav1.ListOptions) (kuberuntime.Object, error) {
			//options.LabelSelector = labelSelector.String()
			return podClient.Pods(namespace).List(context.TODO(), options)
		},
		WatchFunc: func(options metav1.ListOptions) (watch.Interface, error) {
			//options.LabelSelector = labelSelector.String()
			return podClient.Pods(namespace).Watch(context.TODO(), options)
		},
	}

	var result *corev1.Pod
	fmt.Printf("starting the watch\n")
	ev, err := watchtools.UntilWithSync(ctx, lw, &corev1.Pod{}, nil, podRunningAndReady)
	if ev != nil {
		result = ev.Object.(*corev1.Pod)
	}

	return result, err
}

func main() {
	var ()
	watcherStarted := make(chan struct{})
	// Create the fake client.
	client := fake.NewSimpleClientset()

	// A catch-all watch reactor that allows us to inject the watcherStarted channel.
	client.PrependWatchReactor("*", func(action clienttesting.Action) (handled bool, ret watch.Interface, err error) {
		gvr := action.GetResource()
		ns := action.GetNamespace()
		fmt.Printf("starting watch\n")
		watch, err := client.Tracker().Watch(gvr, ns)
		if err != nil {
			return false, nil, err
		}
		close(watcherStarted)
		return true, watch, nil
	})
	client.PrependReactor("list", "pods", func(action clienttesting.Action) (handled bool, ret kuberuntime.Object, err error) {

	} func(action clienttesting.Action) (handled bool, ret watch.Interface, err error) {
		gvr := action.GetResource()
		ns := action.GetNamespace()
		fmt.Printf("starting watch\n")
		watch, err := client.Tracker().Watch(gvr, ns)
		if err != nil {
			return false, nil, err
		}
		close(watcherStarted)
		return true, watch, nil
	})

	pods := make(chan *corev1.Pod, 1)

	go func() {
		pod, err := waitForPod(100 * time.Second)
		fmt.Printf("error: %v\tpod: %v\n", err, pod)
		pods <- pod
	}()

	//// We will create an informer that writes added pods to a channel.
	//informers := informers.NewSharedInformerFactory(client, 0)
	//podInformer := informers.Core().V1().Pods().Informer()
	//podInformer.AddEventHandler(&cache.ResourceEventHandlerFuncs{
	//	AddFunc: func(obj interface{}) {
	//		pod := obj.(*corev1.Pod)
	//		t.Logf("pod added: %s/%s", pod.Namespace, pod.Name)
	//		pods <- pod
	//	},
	//})
	//
	//// Make sure informers are running.
	//informers.Start(ctx.Done())
	//
	//// This is not required in tests, but it serves as a proof-of-concept by
	//// ensuring that the informer goroutine have warmed up and called List before
	//// we send any events to it.
	//cache.WaitForCacheSync(ctx.Done(), podInformer.HasSynced)

	// The fake client doesn't support resource version. Any writes to the client
	// after the informer's initial LIST and before the informer establishing the
	// watcher will be missed by the informer. Therefore we wait until the watcher
	// starts.
	// Note that the fake client isn't designed to work with informer. It
	// doesn't support resource version. It's encouraged to use a real client
	// in an integration/E2E test if you need to test complex behavior with
	// informer/controllers.
	<-watcherStarted
	fmt.Printf("watcher started\n")
	// Inject an event into the fake client.
	p := &corev1.Pod{ObjectMeta: metav1.ObjectMeta{
		Name: "my-pod",
		Labels: map[string]string{
			"app": "istioValidation",
		},
	}}
	_, err := client.CoreV1().Pods("istio-system").Create(context.TODO(), p, metav1.CreateOptions{})
	if err != nil {
		panic(fmt.Errorf("error creating pod: %v\n", err))
	}
	fmt.Println("created pod")

	select {
	case pod := <-pods:
		fmt.Printf("Got pod from channel: %s/%s\n", pod.Namespace, pod.Name)
	case <-time.After(wait.ForeverTestTimeout):
		panic("Informer did not get the added pod")
	}
}
