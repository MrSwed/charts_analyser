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
        "/chart/vessels": {
            "get": {
                "security": [
                    {
                        "BearerAuth": []
                    }
                ],
                "description": "которые пересекали указанные морские карты в заданный временной промежуток.",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "Chart"
                ],
                "summary": "список судов",
                "parameters": [
                    {
                        "description": "Входные параметры: идентификаторы карт, стартовая дата, конечная дата.",
                        "name": "InputZones",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/domain.InputZones"
                        }
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "type": "array",
                            "items": {
                                "type": "integer"
                            }
                        }
                    },
                    "400": {
                        "description": "Bad Request"
                    },
                    "500": {
                        "description": "Internal Server Error"
                    }
                }
            }
        },
        "/chart/zones": {
            "get": {
                "security": [
                    {
                        "BearerAuth": []
                    }
                ],
                "description": "которые пересекались заданными в запросе судами в заданный временной промежуток.",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "Chart"
                ],
                "summary": "список морских карт",
                "parameters": [
                    {
                        "description": "Входные параметры: идентификаторы судов, стартовая дата, конечная дата.",
                        "name": "InputVesselsInterval",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/domain.InputVesselsInterval"
                        }
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "type": "array",
                            "items": {
                                "type": "string"
                            }
                        }
                    },
                    "400": {
                        "description": "Bad Request"
                    },
                    "500": {
                        "description": "Internal Server Error"
                    }
                }
            }
        },
        "/monitor": {
            "get": {
                "security": [
                    {
                        "BearerAuth": []
                    }
                ],
                "description": "поставленных на мониторинг",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "Monitor"
                ],
                "summary": "Список судов",
                "responses": {
                    "200": {
                        "description": "Ok",
                        "schema": {
                            "type": "array",
                            "items": {
                                "$ref": "#/definitions/domain.Vessel"
                            }
                        }
                    },
                    "400": {
                        "description": "Bad Request"
                    },
                    "500": {
                        "description": "Internal Server Error"
                    }
                }
            },
            "post": {
                "security": [
                    {
                        "BearerAuth": []
                    }
                ],
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "Monitor"
                ],
                "summary": "Поставить судно на контроль",
                "parameters": [
                    {
                        "description": "список ID Судов",
                        "name": "VesselIDs",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "type": "array",
                            "items": {
                                "type": "integer"
                            }
                        }
                    }
                ],
                "responses": {
                    "200": {
                        "description": "Ok",
                        "schema": {
                            "type": "string"
                        }
                    },
                    "400": {
                        "description": "Bad Request"
                    },
                    "500": {
                        "description": "Internal Server Error"
                    }
                }
            },
            "delete": {
                "security": [
                    {
                        "BearerAuth": []
                    }
                ],
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "Monitor"
                ],
                "summary": "Снять судно с контроля",
                "parameters": [
                    {
                        "description": "список ID Судов",
                        "name": "VesselIDs",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "type": "array",
                            "items": {
                                "type": "integer"
                            }
                        }
                    }
                ],
                "responses": {
                    "200": {
                        "description": "Ok",
                        "schema": {
                            "type": "string"
                        }
                    },
                    "400": {
                        "description": "Bad Request"
                    },
                    "500": {
                        "description": "Internal Server Error"
                    }
                }
            }
        },
        "/monitor/state": {
            "get": {
                "security": [
                    {
                        "BearerAuth": []
                    }
                ],
                "description": "для выбранных судов, стоящих на мониторинге",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "Monitor"
                ],
                "summary": "Текущие данные",
                "parameters": [
                    {
                        "description": "список ID Судов",
                        "name": "VesselIDs",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "type": "array",
                            "items": {
                                "type": "integer"
                            }
                        }
                    }
                ],
                "responses": {
                    "200": {
                        "description": "Ok",
                        "schema": {
                            "type": "array",
                            "items": {
                                "$ref": "#/definitions/domain.VesselState"
                            }
                        }
                    },
                    "400": {
                        "description": "Bad Request"
                    },
                    "404": {
                        "description": "no data yet"
                    },
                    "500": {
                        "description": "Internal Server Error"
                    }
                }
            }
        },
        "/track": {
            "post": {
                "security": [
                    {
                        "BearerAuth": []
                    }
                ],
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "Track"
                ],
                "summary": "Запись трека судна",
                "parameters": [
                    {
                        "type": "string",
                        "description": "Bearer: JWT claims must have: id key used as vesselID and role: 1",
                        "name": "Authorization",
                        "in": "header",
                        "required": true
                    },
                    {
                        "description": "[lon, lat]",
                        "name": "Point",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "type": "array",
                            "items": {
                                "type": "number"
                            }
                        }
                    }
                ],
                "responses": {
                    "200": {
                        "description": "Ok",
                        "schema": {
                            "type": "string"
                        }
                    },
                    "400": {
                        "description": "Bad Request"
                    },
                    "500": {
                        "description": "Internal Server Error"
                    }
                }
            }
        },
        "/track/{id}": {
            "post": {
                "security": [
                    {
                        "BearerAuth": []
                    }
                ],
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "Track"
                ],
                "summary": "Маршрут судна за указанный период",
                "parameters": [
                    {
                        "type": "integer",
                        "description": "ID Судна ",
                        "name": "id",
                        "in": "path",
                        "required": true
                    },
                    {
                        "type": "string",
                        "name": "finish",
                        "in": "query"
                    },
                    {
                        "type": "string",
                        "name": "start",
                        "in": "query"
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "type": "array",
                            "items": {
                                "$ref": "#/definitions/charts_analyser_internal_app_domain.Track"
                            }
                        }
                    },
                    "400": {
                        "description": "Bad Request"
                    },
                    "500": {
                        "description": "Internal Server Error"
                    }
                }
            }
        },
        "/vessels": {
            "get": {
                "security": [
                    {
                        "BearerAuth": []
                    }
                ],
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "Vessel"
                ],
                "summary": "Добавление судна",
                "parameters": [
                    {
                        "description": "список ID Судов",
                        "name": "VesselNames",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "type": "array",
                            "items": {
                                "type": "array",
                                "items": {
                                    "type": "integer"
                                }
                            }
                        }
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "type": "array",
                            "items": {
                                "$ref": "#/definitions/domain.Vessel"
                            }
                        }
                    },
                    "400": {
                        "description": "Bad Request"
                    },
                    "500": {
                        "description": "Internal Server Error"
                    }
                }
            },
            "post": {
                "security": [
                    {
                        "BearerAuth": []
                    }
                ],
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "Vessel"
                ],
                "summary": "Добавление судна",
                "parameters": [
                    {
                        "description": "список названий Судов",
                        "name": "VesselNames",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "type": "array",
                            "items": {
                                "type": "string"
                            }
                        }
                    }
                ],
                "responses": {
                    "201": {
                        "description": "Created",
                        "schema": {
                            "type": "array",
                            "items": {
                                "$ref": "#/definitions/domain.Vessel"
                            }
                        }
                    },
                    "400": {
                        "description": "Bad Request"
                    },
                    "500": {
                        "description": "Internal Server Error"
                    }
                }
            },
            "delete": {
                "security": [
                    {
                        "BearerAuth": []
                    }
                ],
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "Vessel"
                ],
                "summary": "Удаление судна",
                "parameters": [
                    {
                        "description": "список ID Судов",
                        "name": "VesselNames",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "type": "array",
                            "items": {
                                "type": "array",
                                "items": {
                                    "type": "integer"
                                }
                            }
                        }
                    }
                ],
                "responses": {
                    "200": {
                        "description": "Ok",
                        "schema": {
                            "type": "string"
                        }
                    },
                    "400": {
                        "description": "Bad Request"
                    },
                    "500": {
                        "description": "Internal Server Error"
                    }
                }
            },
            "patch": {
                "security": [
                    {
                        "BearerAuth": []
                    }
                ],
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "Vessel"
                ],
                "summary": "Восстановление судна",
                "parameters": [
                    {
                        "description": "список ID Судов",
                        "name": "VesselNames",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "type": "array",
                            "items": {
                                "type": "array",
                                "items": {
                                    "type": "integer"
                                }
                            }
                        }
                    }
                ],
                "responses": {
                    "200": {
                        "description": "Ok",
                        "schema": {
                            "type": "string"
                        }
                    },
                    "400": {
                        "description": "Bad Request"
                    },
                    "500": {
                        "description": "Internal Server Error"
                    }
                }
            }
        }
    },
    "definitions": {
        "charts_analyser_internal_app_domain.Track": {
            "type": "object",
            "properties": {
                "id": {
                    "type": "integer"
                },
                "location": {
                    "type": "array",
                    "items": {
                        "type": "number"
                    }
                },
                "name": {
                    "type": "string"
                },
                "timestamp": {
                    "type": "string"
                }
            }
        },
        "domain.CurrentZone": {
            "type": "object",
            "properties": {
                "timeIn": {
                    "type": "string"
                },
                "zones": {
                    "type": "array",
                    "items": {
                        "type": "string"
                    }
                }
            }
        },
        "domain.Duration": {
            "type": "integer",
            "enum": [
                -9223372036854775808,
                9223372036854775807,
                1,
                1000,
                1000000,
                1000000000,
                60000000000,
                3600000000000
            ],
            "x-enum-varnames": [
                "minDuration",
                "maxDuration",
                "Nanosecond",
                "Microsecond",
                "Millisecond",
                "Second",
                "Minute",
                "Hour"
            ]
        },
        "domain.InputVesselsInterval": {
            "type": "object",
            "properties": {
                "finish": {
                    "type": "string"
                },
                "start": {
                    "type": "string"
                },
                "vesselIDs": {
                    "type": "array",
                    "items": {
                        "type": "integer"
                    }
                }
            }
        },
        "domain.InputZones": {
            "type": "object",
            "properties": {
                "finish": {
                    "type": "string"
                },
                "start": {
                    "type": "string"
                },
                "zoneNames": {
                    "type": "array",
                    "items": {
                        "type": "string"
                    }
                }
            }
        },
        "domain.Vessel": {
            "type": "object",
            "properties": {
                "id": {
                    "type": "integer"
                },
                "name": {
                    "type": "string"
                }
            }
        },
        "domain.VesselState": {
            "type": "object",
            "properties": {
                "control": {
                    "type": "boolean"
                },
                "controlEnd": {
                    "type": "string"
                },
                "controlStart": {
                    "type": "string"
                },
                "currentZone": {
                    "$ref": "#/definitions/domain.CurrentZone"
                },
                "id": {
                    "type": "integer"
                },
                "location": {
                    "type": "array",
                    "items": {
                        "type": "number"
                    }
                },
                "name": {
                    "type": "string"
                },
                "timestamp": {
                    "type": "string"
                },
                "zoneDuration": {
                    "$ref": "#/definitions/domain.Duration"
                }
            }
        }
    },
    "securityDefinitions": {
        "BearerAuth": {
            "description": "Insert your access token default (Bearer access_token_here)",
            "type": "apiKey",
            "name": "Authorization",
            "in": "header"
        }
    }
}`

// SwaggerInfo holds exported Swagger Info so clients can modify it
var SwaggerInfo = &swag.Spec{
	Version:          "1.0",
	Host:             "localhost:3000",
	BasePath:         "/api/",
	Schemes:          []string{},
	Title:            "Charts analyser: web-service API",
	Description:      "",
	InfoInstanceName: "swagger",
	SwaggerTemplate:  docTemplate,
	LeftDelim:        "{{",
	RightDelim:       "}}",
}

func init() {
	swag.Register(SwaggerInfo.InstanceName(), SwaggerInfo)
}
