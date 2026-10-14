POST   /api/auth/register     {name, email, password}       → {user}
POST   /api/auth/login        {email, password}             → {user}
POST   /api/auth/logout                                     → 204
GET    /api/auth/me                                         → {user}

GET    /api/tests                                           → Test[]
POST   /api/tests             {title, description}          → Test
GET    /api/tests/:id                                       → Test
PATCH  /api/tests/:id         {title, description}          → Test
DELETE /api/tests/:id                                       → 204

POST   /api/tests/:id/questions                             → Question
PATCH  /api/questions/:id                                   → Question
DELETE /api/questions/:id                                   → 204

POST   /api/tests/:id/publish                               → {token, url}
GET    /api/public/tests/:token                             → PublicTest
POST   /api/public/tests/:token/attempts  {name, answers}   → {attemptId, score}
GET    /api/tests/:id/attempts                              → Attempt[]
GET    /api/attempts/:id                                    → AttemptDetail