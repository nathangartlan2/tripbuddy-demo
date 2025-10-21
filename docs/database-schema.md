erDiagram
Park ||--o{ ParkThingToDo : has

      Park {
          int id PK
          string name
          string nps_park_code
          decimal latitude
          decimal longitude
          string state_code
          datetime created_at
          bool is_active
      }

      ParkThingToDo {
          int id PK
          int park_id FK
          string title
          string short_description
          string[] activity_tags
          datetime created_at
          bool is_active
      }
