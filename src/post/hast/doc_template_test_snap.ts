// Jest Snapshot v1, https://goo.gl/fbAQLP

exports[`Doc empty 1`] = `
Object {
  "children": Array [
    Object {
      "name": "html",
      "type": "doctype",
    },
    Object {
      "children": Array [
        Object {
          "children": Array [],
          "tagName": "head",
          "type": "element",
        },
        Object {
          "children": Array [
            Object {
              "children": Array [],
              "tagName": "header",
              "type": "element",
            },
            Object {
              "children": Array [
                Object {
                  "children": Array [],
                  "properties": Object {
                    "className": Array [
                      "main-inner-container",
                    ],
                  },
                  "tagName": "div",
                  "type": "element",
                },
              ],
              "tagName": "main",
              "type": "element",
            },
            Object {
              "children": Array [],
              "tagName": "footer",
              "type": "element",
            },
          ],
          "tagName": "body",
          "type": "element",
        },
      ],
      "properties": Object {
        "lang": "en",
      },
      "tagName": "html",
      "type": "element",
    },
  ],
  "type": "root",
}
`;

exports[`Doc with children 1`] = `
Object {
  "children": Array [
    Object {
      "name": "html",
      "type": "doctype",
    },
    Object {
      "children": Array [
        Object {
          "children": Array [],
          "tagName": "head",
          "type": "element",
        },
        Object {
          "children": Array [
            Object {
              "children": Array [],
              "tagName": "header",
              "type": "element",
            },
            Object {
              "children": Array [
                Object {
                  "children": Array [
                    Object {
                      "children": Array [
                        Object {
                          "type": "text",
                          "value": "alpha",
                        },
                      ],
                      "tagName": "p",
                      "type": "element",
                    },
                  ],
                  "properties": Object {
                    "className": Array [
                      "main-inner-container",
                    ],
                  },
                  "tagName": "div",
                  "type": "element",
                },
              ],
              "tagName": "main",
              "type": "element",
            },
            Object {
              "children": Array [],
              "tagName": "footer",
              "type": "element",
            },
          ],
          "tagName": "body",
          "type": "element",
        },
      ],
      "properties": Object {
        "lang": "en",
      },
      "tagName": "html",
      "type": "element",
    },
  ],
  "type": "root",
}
`;

exports[`Doc with title 1`] = `
Object {
  "children": Array [
    Object {
      "name": "html",
      "type": "doctype",
    },
    Object {
      "children": Array [
        Object {
          "children": Array [
            Object {
              "children": Array [
                Object {
                  "type": "text",
                  "value": "alpha",
                },
              ],
              "tagName": "title",
              "type": "element",
            },
          ],
          "tagName": "head",
          "type": "element",
        },
        Object {
          "children": Array [
            Object {
              "children": Array [],
              "tagName": "header",
              "type": "element",
            },
            Object {
              "children": Array [
                Object {
                  "children": Array [],
                  "properties": Object {
                    "className": Array [
                      "main-inner-container",
                    ],
                  },
                  "tagName": "div",
                  "type": "element",
                },
              ],
              "tagName": "main",
              "type": "element",
            },
            Object {
              "children": Array [],
              "tagName": "footer",
              "type": "element",
            },
          ],
          "tagName": "body",
          "type": "element",
        },
      ],
      "properties": Object {
        "lang": "en",
      },
      "tagName": "html",
      "type": "element",
    },
  ],
  "type": "root",
}
`;
