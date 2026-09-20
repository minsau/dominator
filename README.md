# Dominator

Dominator es el registrar dinámico de rutas para Traefik en el homelab (`192.168.0.108`).

Provee una API HTTP que actúa como provider `http` para Traefik, sirviendo dinámicamente configuraciones de routers y services desde la base de datos `dominator_db` en PostgreSQL (`192.168.0.214:5432`).

## Arquitectura

- **Store**: PostgreSQL (`dominator_db.routes`).
- **Co-locación**: Corre en `.108` en el mismo docker-compose que Traefik con `depends_on: condition: service_healthy`.
- **Endpoints**:
  - `GET /traefik-config`: Devuelve configuración en formato JSON para Traefik Dynamic HTTP Provider.
  - `GET /health`: Chequea estado de la conexión a DB y conteo de rutas (`routes_served`).
