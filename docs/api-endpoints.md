## Parks

| Method | Endpoint                | Description        | Auth |
| ------ | ----------------------- | ------------------ | ---- |
| GET    | /api/parks              | Get all parks      | No   |
| GET    | /api/parks/{id}         | Get park by ID     | No   |
| GET    | /api/parks/state/{code} | Get parks by state | No   |

## Things To Do

| Method | Endpoint                         | Description             | Auth |
| ------ | -------------------------------- | ----------------------- | ---- |
| GET    | /api/parks/{parkId}/things-to-do | Get activities for park | No   |
| GET    | /api/thingstodo/{id}             | Get activity by ID      | No   |
