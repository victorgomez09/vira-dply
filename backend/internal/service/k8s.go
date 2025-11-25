package service

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/victorgomez09/vira-dply/internal/utils"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/utils/ptr"
)

type K8sClient struct {
	client *kubernetes.Clientset
}

func NewK8sClient() (*K8sClient, error) {
	var config *rest.Config
	var err error

	// 1. INTENTO: Configuración In-Cluster (Para producción/ejecución en K8s)
	config, err = rest.InClusterConfig()
	if err == nil {
		log.Println("Usando configuración In-Cluster de Kubernetes.")
	} else {
		// 2. FALLBACK: Configuración Kubeconfig local (Para desarrollo)
		log.Println("Fallo al obtener In-Cluster config. Intentando Kubeconfig local...")

		// A. Leer la ruta de la variable de entorno personalizada
		kubeconfigPath := os.Getenv("KUBECONFIG_PATH")

		if kubeconfigPath == "" {
			// B. Si no se define KUBECONFIG_PATH, usar la ruta estándar (~/.kube/config)
			kubeconfigPath = filepath.Join(os.Getenv("HOME"), ".kube", "config")
		}

		// Cargar la configuración desde el archivo kubeconfig
		config, err = clientcmd.BuildConfigFromFlags("", kubeconfigPath)
		if err != nil {
			// Si falla la configuración local, lanzamos el error crítico
			return nil, fmt.Errorf("fallo al obtener configuración In-Cluster y Kubeconfig local: %w", err)
		}
		log.Printf("Usando configuración de Kubeconfig en: %s", kubeconfigPath)
	}

	// Usar la configuración obtenida para crear el cliente
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("fallo al crear el cliente K8s: %w", err)
	}

	return &K8sClient{client: clientset}, nil
}

func (c *K8sClient) CreateNamespace(ctx context.Context, namespaceName string) error {
	nsSpec := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{Name: namespaceName},
	}
	_, err := c.client.CoreV1().Namespaces().Create(ctx, nsSpec, metav1.CreateOptions{})
	if err != nil && !IsAlreadyExistsError(err) {
		return fmt.Errorf("fallo al crear Namespace: %w", err)
	}

	return nil
}

func (c *K8sClient) CreateDeployment(ctx context.Context, namespace, appName, imageName string) error {
	log.Printf("Creating new deployment %s", appName)
	manager := NewRegistryManager()
	pullSecretName := os.Getenv("REGISTRY_PULL_SECRET_NAME")
	// registryHost := os.Getenv("K8S_REGISTRY_SERVICE_HOST")
	registryHost := utils.GetOutboundIP() + ":5000"
	fmt.Printf("registry host %s", registryHost)

	secretData, err := manager.CreateK8sPullSecretData(registryHost)
	if err != nil {
		return fmt.Errorf("fallo al crear datos del ImagePullSecret: %w", err)
	}

	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{Name: pullSecretName, Namespace: namespace},
		Type:       corev1.SecretTypeDockerConfigJson,
		Data:       secretData,
	}

	_, err = c.client.CoreV1().Secrets(namespace).Create(ctx, secret, metav1.CreateOptions{})
	if err != nil && !apierrors.IsAlreadyExists(err) {
		return fmt.Errorf("fallo al crear ImagePullSecret: %w", err)
	}

	appName = utils.CleanString(appName)
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
	err = c.createOrUpdateDeployment(ctx, namespace, appName, deployment)
	if err != nil {
		return fmt.Errorf("fallo al crear/actualizar Deployment: %w", err)
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
	_, err = c.client.CoreV1().Services(namespace).Get(ctx, appName, metav1.GetOptions{})

	if err != nil && apierrors.IsNotFound(err) {
		// Service NO existe -> Crear
		log.Printf("Service %s no encontrado, creando nuevo.", appName)
		_, err = c.client.CoreV1().Services(namespace).Create(ctx, service, metav1.CreateOptions{})
		if err != nil {
			return fmt.Errorf("fallo al crear Service: %w", err)
		}
	} else if err == nil {
		// Service SÍ existe -> Ignorar (Los Services generalmente no cambian en este esquema)
		log.Printf("Service %s ya existe, continuando.", appName)
	} else {
		// Otro error de Get
		return fmt.Errorf("fallo al obtener el Service existente: %w", err)
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
	existingIngress, err := c.client.NetworkingV1().Ingresses(namespace).Get(ctx, appName, metav1.GetOptions{})
	if err != nil && apierrors.IsNotFound(err) {
		// Ingress NO existe -> Crear
		log.Printf("Ingress %s no encontrado, creando nuevo.", appName)
		_, err = c.client.NetworkingV1().Ingresses(namespace).Create(ctx, ingress, metav1.CreateOptions{})
		if err != nil {
			return fmt.Errorf("fallo al crear Ingress: %w", err)
		}
	} else if err == nil {
		// Ingress SÍ existe -> Actualizar
		log.Printf("Ingress %s encontrado, actualizando.", appName)

		// Preservar metadatos y actualizar Spec y Annotations
		ingress.ObjectMeta.ResourceVersion = existingIngress.ObjectMeta.ResourceVersion
		ingress.ObjectMeta.UID = existingIngress.ObjectMeta.UID
		ingress.ObjectMeta.Annotations = existingIngress.ObjectMeta.Annotations // Preservar anotaciones existentes

		// Copiar las reglas y especificaciones de red al objeto existente
		existingIngress.Spec = ingress.Spec
		existingIngress.ObjectMeta.Annotations = ingress.ObjectMeta.Annotations

		_, err = c.client.NetworkingV1().Ingresses(namespace).Update(ctx, existingIngress, metav1.UpdateOptions{})
		if err != nil {
			return fmt.Errorf("fallo al actualizar Ingress: %w", err)
		}
	} else {
		// Otro error de Get
		return fmt.Errorf("fallo al obtener el Ingress existente: %w", err)
	}

	return nil
}

// IsAlreadyExistsError es una función de utilidad para manejar errores.
func IsAlreadyExistsError(err error) bool {
	return apierrors.IsAlreadyExists(err)
}

// createOrUpdateDeployment intenta obtener el Deployment; si existe, lo actualiza, si no, lo crea.
func (c *K8sClient) createOrUpdateDeployment(ctx context.Context, namespace, appName string, newDeployment *appsv1.Deployment) error {

	// 1. Intentar obtener el Deployment existente
	existingDeployment, err := c.client.AppsV1().Deployments(namespace).Get(ctx, appName, metav1.GetOptions{})

	if err != nil {
		if apierrors.IsNotFound(err) {
			// El Deployment NO existe -> Crear
			log.Printf("Deployment %s no encontrado, creando nuevo.", appName)
			_, createErr := c.client.AppsV1().Deployments(namespace).Create(ctx, newDeployment, metav1.CreateOptions{})
			return createErr
		}
		// Otro error de Get (permisos, conexión, etc.)
		return fmt.Errorf("fallo al obtener el Deployment existente: %w", err)
	}

	// 2. El Deployment SÍ existe -> Actualizar
	log.Printf("Deployment %s encontrado, actualizando.", appName)

	// Preservar la metadata necesaria (UID, ResourceVersion) y actualizar solo la Spec.
	newDeployment.ObjectMeta.ResourceVersion = existingDeployment.ObjectMeta.ResourceVersion
	newDeployment.ObjectMeta.UID = existingDeployment.ObjectMeta.UID

	// NOTA: Es crucial que el Selector de la Spec sea inmutable.
	// Aquí actualizamos la Spec.Template (PodTemplate) y Replicas (si las hubiera).
	existingDeployment.Spec.Template = newDeployment.Spec.Template
	existingDeployment.Spec.Replicas = newDeployment.Spec.Replicas // Si Replicas está en newDeployment

	_, updateErr := c.client.AppsV1().Deployments(namespace).Update(ctx, existingDeployment, metav1.UpdateOptions{})
	return updateErr
}
