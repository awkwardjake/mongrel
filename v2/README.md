# mongrel

[MongoDB Go Driver](https://www.mongodb.com/docs/drivers/go/current/) dependent connect/disconnect package

## Install

```text
go get github.com/awkwardjake/mongrel/v2
```

## Docker Compose example

Example MongoDB service configuration for your docker-compose.yml file

```yaml
version: "3.9"

services:
  mongo:
    image: mongo
    container_name: backendDB
    env_file:
      - .env
    restart: always
    ports:
      - "${MONGODBPORT}:27017"
    environment:
      MONGO_INITDB_DATABASE: ${DBAPP}
      MONGO_INITDB_ROOT_USERNAME: ${ROOTDBUSER}
      MONGO_INITDB_ROOT_PASSWORD: ${ROOTDBPASSWORD}
    volumes:
      - ./scripts/userScript.js:/docker-entrypoint-initdb.d/user.js:ro
      - mongodb_data_container:/usr/apps/exampleDB/database/mongo/db

volumes:
  mongodb_data_container:
```

### [Create a user](https://www.mongodb.com/docs/manual/tutorial/create-users/) JavaScript

Add this user create JavaScript file to project directory and reference it in docker-compose.yml. If using example above, it would be `./scripts/userScript.js`

Modify JS to suit needs in terms of user credentials and permissions within MongoDB

```javascript
db.createUser(
    {
        user:"appUser",
        pwd:"appUserPass",
        roles: 
        [
            {
                role: "readWrite", 
                db: "test"
            }
        ]
    }
)
```
