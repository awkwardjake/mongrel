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