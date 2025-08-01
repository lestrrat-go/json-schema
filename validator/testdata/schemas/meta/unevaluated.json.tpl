{
    "$schema": "{{.ServerURL}}/draft/2020-12/schema",
    "$id": "{{.ServerURL}}/draft/2020-12/meta/unevaluated",
    "$dynamicAnchor": "meta",

    "title": "Unevaluated applicator vocabulary meta-schema",
    "type": ["object", "boolean"],
    "properties": {
        "unevaluatedItems": { "$dynamicRef": "#meta" },
        "unevaluatedProperties": { "$dynamicRef": "#meta" }
    }
}
