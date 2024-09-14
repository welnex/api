// Package docs Code generated by swaggo/swag. DO NOT EDIT
package docs

import "github.com/swaggo/swag"

const docTemplate = `{
    "schemes": {{ marshal .Schemes }},
    "swagger": "2.0",
    "info": {
        "description": "{{escape .Description}}",
        "title": "{{.Title}}",
        "contact": {},
        "version": "{{.Version}}"
    },
    "host": "{{.Host}}",
    "basePath": "{{.BasePath}}",
    "paths": {
        "/activity": {
            "get": {
                "security": [
                    {
                        "AccessToken": []
                    }
                ],
                "description": "reeturns user's workout history",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "workout"
                ],
                "summary": "Get user's activity history",
                "parameters": [
                    {
                        "type": "string",
                        "description": "Begin date",
                        "name": "begin",
                        "in": "query",
                        "required": true
                    },
                    {
                        "type": "string",
                        "description": "End date",
                        "name": "end",
                        "in": "query",
                        "required": true
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/responsebody.ActivityHistory"
                        }
                    },
                    "400": {
                        "description": "Bad Request",
                        "schema": {
                            "$ref": "#/definitions/responsebody.Message"
                        }
                    }
                }
            }
        },
        "/auth/account": {
            "get": {
                "security": [
                    {
                        "AccessToken": []
                    }
                ],
                "description": "returns an user's information, that currently logged in",
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "user"
                ],
                "summary": "Get information about current user",
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/responsebody.User"
                        }
                    },
                    "401": {
                        "description": "Unauthorized",
                        "schema": {
                            "$ref": "#/definitions/responsebody.Message"
                        }
                    }
                }
            },
            "post": {
                "description": "create user in database",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "auth"
                ],
                "summary": "Create new account",
                "parameters": [
                    {
                        "description": "User information",
                        "name": "input",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/requestbody.CreateAccount"
                        }
                    }
                ],
                "responses": {
                    "201": {
                        "description": "Created",
                        "schema": {
                            "$ref": "#/definitions/responsebody.User"
                        }
                    },
                    "400": {
                        "description": "Bad Request",
                        "schema": {
                            "$ref": "#/definitions/responsebody.Message"
                        }
                    },
                    "409": {
                        "description": "Conflict",
                        "schema": {
                            "$ref": "#/definitions/responsebody.Message"
                        }
                    }
                }
            }
        },
        "/auth/account/confirm": {
            "post": {
                "description": "confirms user's email",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "auth"
                ],
                "summary": "Confirm account's email",
                "parameters": [
                    {
                        "description": "Token",
                        "name": "input",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/requestbody.ConfirmAccount"
                        }
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK"
                    },
                    "400": {
                        "description": "Bad Request",
                        "schema": {
                            "$ref": "#/definitions/responsebody.Message"
                        }
                    },
                    "404": {
                        "description": "Not Found",
                        "schema": {
                            "$ref": "#/definitions/responsebody.Message"
                        }
                    }
                }
            }
        },
        "/auth/password": {
            "patch": {
                "description": "updates password for user",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "auth"
                ],
                "summary": "Update password",
                "parameters": [
                    {
                        "description": "User information",
                        "name": "input",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/requestbody.UpdatePassword"
                        }
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK"
                    },
                    "400": {
                        "description": "Bad Request",
                        "schema": {
                            "$ref": "#/definitions/responsebody.Message"
                        }
                    },
                    "404": {
                        "description": "Not Found",
                        "schema": {
                            "$ref": "#/definitions/responsebody.Message"
                        }
                    }
                }
            }
        },
        "/auth/password/reset": {
            "post": {
                "description": "sends an email with recovery link",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "auth"
                ],
                "summary": "Request password reset",
                "parameters": [
                    {
                        "description": "User information",
                        "name": "input",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/requestbody.ResetPassword"
                        }
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK"
                    },
                    "400": {
                        "description": "Bad Request",
                        "schema": {
                            "$ref": "#/definitions/responsebody.Message"
                        }
                    },
                    "404": {
                        "description": "Not Found",
                        "schema": {
                            "$ref": "#/definitions/responsebody.Message"
                        }
                    }
                }
            }
        },
        "/auth/session": {
            "post": {
                "description": "check if user exists, and return an access token",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "auth"
                ],
                "summary": "Create a session for existing account",
                "parameters": [
                    {
                        "description": "User information",
                        "name": "input",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/requestbody.CreateSession"
                        }
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/responsebody.Token"
                        }
                    },
                    "400": {
                        "description": "Bad Request",
                        "schema": {
                            "$ref": "#/definitions/responsebody.Message"
                        }
                    },
                    "401": {
                        "description": "Unauthorized",
                        "schema": {
                            "$ref": "#/definitions/responsebody.Message"
                        }
                    }
                }
            }
        },
        "/healthcheck": {
            "get": {
                "description": "check if server status is ok",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "status"
                ],
                "summary": "Ping a server",
                "responses": {
                    "200": {
                        "description": "ok",
                        "schema": {
                            "type": "string"
                        }
                    }
                }
            }
        },
        "/workout": {
            "post": {
                "security": [
                    {
                        "AccessToken": []
                    }
                ],
                "description": "creates a new record about workout session",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "workout"
                ],
                "summary": "Create a record about past workout",
                "parameters": [
                    {
                        "description": "Information about workout session",
                        "name": "input",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/requestbody.CreateWorkout"
                        }
                    }
                ],
                "responses": {
                    "201": {
                        "description": "Created",
                        "schema": {
                            "$ref": "#/definitions/responsebody.Workout"
                        }
                    },
                    "400": {
                        "description": "Bad Request",
                        "schema": {
                            "$ref": "#/definitions/responsebody.Message"
                        }
                    }
                }
            }
        }
    },
    "definitions": {
        "requestbody.ConfirmAccount": {
            "type": "object",
            "required": [
                "token"
            ],
            "properties": {
                "token": {
                    "type": "string"
                }
            }
        },
        "requestbody.CreateAccount": {
            "type": "object",
            "required": [
                "email",
                "password",
                "username"
            ],
            "properties": {
                "email": {
                    "type": "string"
                },
                "password": {
                    "type": "string"
                },
                "username": {
                    "type": "string"
                }
            }
        },
        "requestbody.CreateSession": {
            "type": "object",
            "required": [
                "email",
                "password"
            ],
            "properties": {
                "email": {
                    "type": "string"
                },
                "password": {
                    "type": "string"
                }
            }
        },
        "requestbody.CreateWorkout": {
            "type": "object",
            "required": [
                "date",
                "duration",
                "kind"
            ],
            "properties": {
                "date": {
                    "type": "string"
                },
                "duration": {
                    "type": "integer"
                },
                "kind": {
                    "type": "string"
                }
            }
        },
        "requestbody.ResetPassword": {
            "type": "object",
            "required": [
                "email"
            ],
            "properties": {
                "email": {
                    "type": "string"
                }
            }
        },
        "requestbody.UpdatePassword": {
            "type": "object",
            "required": [
                "password",
                "token"
            ],
            "properties": {
                "password": {
                    "type": "string"
                },
                "token": {
                    "type": "string"
                }
            }
        },
        "responsebody.ActivityHistory": {
            "type": "object",
            "properties": {
                "count": {
                    "type": "integer"
                },
                "user_id": {
                    "type": "string"
                },
                "workouts": {
                    "type": "array",
                    "items": {
                        "$ref": "#/definitions/responsebody.Workout"
                    }
                }
            }
        },
        "responsebody.Message": {
            "type": "object",
            "properties": {
                "message": {
                    "type": "string"
                }
            }
        },
        "responsebody.Token": {
            "type": "object",
            "properties": {
                "token": {
                    "type": "string"
                }
            }
        },
        "responsebody.User": {
            "type": "object",
            "properties": {
                "display_name": {
                    "type": "string"
                },
                "email": {
                    "type": "string"
                },
                "id": {
                    "type": "string"
                },
                "username": {
                    "type": "string"
                }
            }
        },
        "responsebody.Workout": {
            "type": "object",
            "properties": {
                "date": {
                    "type": "string"
                },
                "duration": {
                    "type": "integer"
                },
                "id": {
                    "type": "string"
                },
                "kind": {
                    "type": "string"
                }
            }
        }
    },
    "securityDefinitions": {
        "AccessToken": {
            "type": "apiKey",
            "name": "Authorization",
            "in": "header"
        }
    }
}`

// SwaggerInfo holds exported Swagger Info so clients can modify it
var SwaggerInfo = &swag.Spec{
	Version:          "0.1",
	Host:             "localhost:6969",
	BasePath:         "/api",
	Schemes:          []string{},
	Title:            "Welnex API",
	Description:      "API server for Welnex application",
	InfoInstanceName: "swagger",
	SwaggerTemplate:  docTemplate,
	LeftDelim:        "{{",
	RightDelim:       "}}",
}

func init() {
	swag.Register(SwaggerInfo.InstanceName(), SwaggerInfo)
}
