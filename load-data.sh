#!/bin/bash

events_source_file="https://github.com/hackload-kz/data/releases/download/2025-08-15/events.sql"
users_source_file="https://github.com/hackload-kz/data/releases/download/2025-08-15/users.sql"

curl -o ./data/events.sql $events_source_file
curl -o ./data/users.sql $users_source_file

psql -h localhost -p 5432 -U bulbul -d bulbul -f data/users.sql
psql -h localhost -p 5432 -U bulbul -d bulbul -f data/events.sql