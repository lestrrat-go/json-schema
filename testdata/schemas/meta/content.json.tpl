{
    "$schema": "{{.ServerURL}}/draft/2020-12/schema",
    "$id": "{{.ServerURL}}/draft/2020-12/meta/content",
    "$dynamicAnchor": "meta",

    "title": "Content vocabulary meta-schema",

    "type": ["object", "boolean"],
    "properties": {
        "contentEncoding": { "type": "string" },
        "contentMediaType": { "type": "string" },
        "contentSchema": { "$dynamicRef": "#meta" }
    }
}
