package model

import "time"

type ServiceType string

const (
	ServicePostgres ServiceType = "postgres"
	ServiceRedis    ServiceType = "redis"
	ServiceMySQL    ServiceType = "mysql"
	ServiceMongo    ServiceType = "mongo"
	ServiceMinio    ServiceType = "minio"
)

type Service struct {
	ID          string      `json:"id"`
	Name        string      `json:"name"`
	Type        ServiceType `json:"type"`
	Image       string      `json:"image"`
	Status      string      `json:"status"` // running | stopped | error
	ContainerID string      `json:"container_id,omitempty"`
	Config      string      `json:"config,omitempty"` // JSON: ports, volumes, env
	CreatedAt   time.Time   `json:"created_at"`
}

type CreateServiceInput struct {
	Name  string      `json:"name"`
	Type  ServiceType `json:"type"`
	Image string      `json:"image,omitempty"` // optional override
}

type ServiceLink struct {
	ID        string    `json:"id"`
	ProjectID string    `json:"project_id"`
	ServiceID string    `json:"service_id"`
	EnvPrefix string    `json:"env_prefix"` // e.g. DATABASE_ → DATABASE_URL, DATABASE_HOST
	CreatedAt time.Time `json:"created_at"`
}

// ServiceDefaults returns default image and config for a service type.
func ServiceDefaults(svcType ServiceType) (image string, env map[string]string, port string) {
	switch svcType {
	case ServicePostgres:
		return "postgres:16-alpine", map[string]string{
			"POSTGRES_USER":     "mypaas",
			"POSTGRES_PASSWORD": "mypaas",
			"POSTGRES_DB":       "mypaas",
		}, "5432"
	case ServiceRedis:
		return "redis:7-alpine", nil, "6379"
	case ServiceMySQL:
		return "mysql:8", map[string]string{
			"MYSQL_ROOT_PASSWORD": "mypaas",
			"MYSQL_DATABASE":      "mypaas",
		}, "3306"
	case ServiceMongo:
		return "mongo:7", map[string]string{
			"MONGO_INITDB_ROOT_USERNAME": "mypaas",
			"MONGO_INITDB_ROOT_PASSWORD": "mypaas",
		}, "27017"
	case ServiceMinio:
		return "minio/minio:latest", map[string]string{
			"MINIO_ROOT_USER":     "minioadmin",
			"MINIO_ROOT_PASSWORD": "minioadmin",
		}, "9000"
	default:
		return "", nil, ""
	}
}

// ServiceConnectionEnv returns env vars to inject into linked projects.
func ServiceConnectionEnv(svc *Service, prefix string) map[string]string {
	containerHost := "mypaas-svc-" + svc.Name
	switch svc.Type {
	case ServicePostgres:
		return map[string]string{
			prefix + "URL":  "postgres://mypaas:mypaas@" + containerHost + ":5432/mypaas?sslmode=disable",
			prefix + "HOST": containerHost,
			prefix + "PORT": "5432",
		}
	case ServiceRedis:
		return map[string]string{
			prefix + "URL":  "redis://" + containerHost + ":6379",
			prefix + "HOST": containerHost,
			prefix + "PORT": "6379",
		}
	case ServiceMySQL:
		return map[string]string{
			prefix + "URL":  "mysql://root:mypaas@" + containerHost + ":3306/mypaas",
			prefix + "HOST": containerHost,
			prefix + "PORT": "3306",
		}
	case ServiceMongo:
		return map[string]string{
			prefix + "URL":  "mongodb://mypaas:mypaas@" + containerHost + ":27017",
			prefix + "HOST": containerHost,
			prefix + "PORT": "27017",
		}
	default:
		return nil
	}
}
