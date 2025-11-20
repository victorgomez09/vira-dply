package service

import (
	"context"
	"fmt"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/utils/ptr"
)

type K8sClient struct {
	client *kubernetes.Clientset
}

func NewK8sClient() (*K8sClient, error) {
	// 1. In-cluster configuration(if controller run inside k8s cluster)
	config, err := rest.InClusterConfig()
	if err != nil {
		// 2. Outside cluster configuration(for local development)
		// If InClusterConfig fails, attempt to load the configuration from Kubeconfig
		// Uncomment and adjust if running outside the cluster:
		// home := os.Getenv("HOME")
		// kubeconfigPath := filepath.Join(home, ".kube", "config")
		// configuration, err = clientcmd.BuildConfigFromFlags("", kubeconfigPath)
		// if error != null {
		// return nil, fmt.Errorf("Failed to get K8s configuration: %w", err)
		// }
		// For now, we'll assume InClusterConfig fails in dev and is successful in prod.
		// In production, use ONLY rest.InClusterConfig()
		return nil, fmt.Errorf("error obtaining in-cluster config: %w", err)
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("error creating k8s client: %w", err)
	}

	return &K8sClient{client: clientset}, nil
}

func (c *K8sClient) CreateDeployment(ctx context.Context, namespace, appName, imageName string) error {
	// Crear Namespace
	nsSpec := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{Name: namespace},
	}
	_, err := c.client.CoreV1().Namespaces().Create(ctx, nsSpec, metav1.CreateOptions{})
	if err != nil && !IsAlreadyExistsError(err) {
		return fmt.Errorf("fallo al crear Namespace: %w", err)
	}

	// 1. Crear Deployment
	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      appName,
			Namespace: namespace,
		},
		Spec: appsv1.DeploymentSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{"app": appName},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{"app": appName},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  appName,
							Image: imageName,
							Ports: []corev1.ContainerPort{
								{ContainerPort: 8080}, // Puerto asumido por defecto para la mayoría de los builds
							},
						},
					},
				},
			},
		},
	}
	// TODO: Considerar aplicar en lugar de crear (CreateOrUpdate)
	_, err = c.client.AppsV1().Deployments(namespace).Create(ctx, deployment, metav1.CreateOptions{})
	if err != nil {
		return fmt.Errorf("fallo al crear Deployment: %w", err)
	}

	// 2. Crear Service (ClusterIP)
	service := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{Name: appName, Namespace: namespace},
		Spec: corev1.ServiceSpec{
			Selector: map[string]string{"app": appName},
			Ports:    []corev1.ServicePort{{Port: 80, TargetPort: intstr.FromInt(8080)}},
			Type:     corev1.ServiceTypeClusterIP,
		},
	}
	_, err = c.client.CoreV1().Services(namespace).Create(ctx, service, metav1.CreateOptions{})
	if err != nil {
		return fmt.Errorf("fallo al crear Service: %w", err)
	}

	// 3. Crear Ingress
	ingress := &networkingv1.Ingress{
		ObjectMeta: metav1.ObjectMeta{
			Name:      appName,
			Namespace: namespace,
			Annotations: map[string]string{
				// Ejemplo de anotación para Cert-Manager (TLS automático)
				"cert-manager.io/cluster-issuer": "letsencrypt-prod",
			},
		},
		Spec: networkingv1.IngressSpec{
			Rules: []networkingv1.IngressRule{
				{
					Host: fmt.Sprintf("%s.localhost", appName),
					IngressRuleValue: networkingv1.IngressRuleValue{
						HTTP: &networkingv1.HTTPIngressRuleValue{
							Paths: []networkingv1.HTTPIngressPath{
								{
									Path:     "/",
									PathType: (*networkingv1.PathType)(ptr.To(networkingv1.PathTypePrefix)),
									Backend: networkingv1.IngressBackend{
										Service: &networkingv1.IngressServiceBackend{
											Name: appName,
											Port: networkingv1.ServiceBackendPort{Number: 80},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}
	_, err = c.client.NetworkingV1().Ingresses(namespace).Create(ctx, ingress, metav1.CreateOptions{})
	if err != nil {
		return fmt.Errorf("fallo al crear Ingress: %w", err)
	}

	return nil
}

// IsAlreadyExistsError es una función de utilidad para manejar errores.
func IsAlreadyExistsError(err error) bool {
	return apierrors.IsAlreadyExists(err)
}
