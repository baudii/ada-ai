You will receive two blocks:

1. A Go handler file containing a handler function with a placeholder comment and all the context about this handler that you'll need as Go comment blocks:
--- HANDLER_FILE_START ---
   <go file>
--- HANDLER_FILE_END ---

2. Existing context (may be `none`) about interfaces, models, and the Server struct:
--- EXISTING_CONTEXT_START ---
   <context>
--- EXISTING_CONTEXT_END ---

**Task:** Replace the placeholder in a handler file by generating **only the handler function body**.

* Don’t output the function signature or any unrelated code.
* Write valid, idiomatic Go (imports handled elsewhere).

**Dependencies / interfaces:**
When the handler needs an external call for `data` or `business logic` you should perform this call using an interface of a current resource's service. Choose 1:
1. Use an existing interface + method, or
2. Extend an existing interface with new method(s), or
3. Create a new interface if nothing fits.

* Assume an interface instance to be a field of a `Server` struct.
* Prefer existing interfaces/models. Only extend/create new ones if no existing method fits. Do not create duplicates.

**If you add anything new (Server fields, interfaces, methods, or models):**

* List new `Server` fields in `## added_fields`.
* Put full definitions of **new or extended** interfaces in `## interfaces` (don’t repeat unchanged ones).
* Put full definitions of any new models in `## added_models`.
* Prefix new models with `models.` inside the handler function body. DON'T PREFIX inside the `## added_models` section.
* Describe every interface you output in `## interfaces_description` using the required structured format.

**Output exactly five sections, in order:**

1. `## function` — one Go code block with **only** the body of the function without signature and open/close brackets, wrapped as a markdown go code block.
2. `## added_fields` — A complete definition of the field: `- FieldName FieldType [Optional: Tag, Comment]` per line, all space separated or `none`.
3. `## interfaces` — one Go code block listing all new/extended interfaces including definition and all methods, or `none`.
4. `## added_models` — one Go code block listing all new models, including definition and all fields, or `none`.
5. `## interfaces_description` — structured descriptions for interfaces in section 3, or `none`. The structure must follow this format:
```
Interface: <Name>
Description: <description of what this interface represents and why it is needed>

Methods:
  - <MethodName>
    Signature: <full Go method signature>
    Description: <what the method does and why it is needed>
```

--- HANDLER_FILE_START ---
%v
--- HANDLER_FILE_END ---

--- EXISTING_CONTEXT_START ---
%v
--- EXISTING_CONTEXT_END ---