# mongrel

Mongo connect/disconnect package

## Install

```text
go get github.com/awkwardjake/mongrel
```

## Docker Compose example

```yaml
version: "3.9"

services:
  mongo:
    image: mongo
    restart: always
    ports:
      - "27017:27017"
    environment:
      MONGO_INITDB_DATABASE: exampleDB
      MONGO_INITDB_ROOT_USERNAME: root
      MONGO_INITDB_ROOT_PASSWORD: rootPass
    volumes:
      - ./scripts/userScript.js:/docker-entrypoint-initdb.d/user.js:ro
      - mongodb_data_container:/usr/apps/exampleDB/database/mongo/db

volumes:
  mongodb_data_container:
```

### user javascript

```javascript
db.createUser(
    {
        user:"appUser",
        pwd:"appUserPass",
        roles: 
        [
            {
                role: "readWrite", 
                db: "exampleDB"
            }
        ]
    }
)
```
