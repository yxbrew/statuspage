package router

const swaggerDoc = `{
  "openapi": "3.0.3",
  "info": {
    "title": "Status Page Backend API",
    "description": "Common HTTP router API for status page backend",
    "version": "1.0.0"
  },
  "servers": [
    {
      "url": "http://localhost:8080"
    }
  ],
  "paths": {
    "/api/v1/health": {
      "get": {
        "summary": "Get service health",
        "operationId": "getHealth",
        "responses": {
          "200": {
            "description": "Service health",
            "content": {
              "application/json": {
                "schema": {
                  "type": "object",
                  "properties": {
                    "status": {
                      "type": "string",
                      "example": "ok"
                    },
                    "service": {
                      "type": "string",
                      "example": "statuspage-backend"
                    },
                    "version": {
                      "type": "string",
                      "example": "1.0.0"
                    }
                  }
                }
              }
            }
          }
        }
      }
    }
  }
}`
