You will receive **one handler file** and **optional existing context blocks**.
Context blocks may be `none`, and may include any of: existing interfaces, existing models, and the Server struct. Treat all provided context as the single source of truth.

**Input format**

1. Handler file (always present):
   --- HANDLER_FILE_START ---

   <go file with a single handler function containing a placeholder comment>

--- HANDLER_FILE_END ---

2. Existing interfaces (may be `none`):
   --- EXISTING_INTERFACES_START ---

   <interfaces>

--- EXISTING_INTERFACES_END ---

3. Existing models (may be `none`):
   --- EXISTING_MODELS_START ---

   <models>

--- EXISTING_MODELS_END ---

4. Server struct (may be `none`):
   --- SERVER_STRUCT_START ---

   <server struct>

--- SERVER_STRUCT_END ---

---

## Task

Replace the placeholder inside the handler by generating **only the handler function body** (no signature, no braces, no other code).
Write valid, idiomatic Go. Imports are handled elsewhere.

**Important:** Your response must follow indicated output format with **exactly five sections** (function body + metadata).

---

## Dependency / interface rules (strict priority)

When the handler needs storage / business logic / external calls:

1. **Prefer existing interfaces and methods** from context.
2. If no existing method fits, **extend an existing interface** with the minimal new method(s).
3. Only if nothing fits, **create a new interface**.

**Never create duplicates** of existing interfaces/models/methods.
Assume any interface you use is a field on `Server`.
If you need a new `Server` field, list it in `## added_fields`.

---

## Coherence, minimal diff, and conventions

* Make **minimal, localized logic changes**: do not rewrite unrelated parts.
* **Preserve conventions** seen in context (naming, error style, response format, status codes).
* Use `ctx := r.Context()` for all downstream calls.
* Do not add new imports or change function signature.

---

## Defaults when conventions are missing

If existing patterns are not discoverable in context, use these defaults:

* **Success responses:**

  * `GET` handlers: `200 OK` with JSON of the result.
  * `POST` handlers: `201 Created` with JSON of created entity (if applicable).
* **Error responses:**

  * Use appropriate status (`400/401/403/404/409/500`).
  * Return a minimal JSON envelope: `{"error": "<message>"}`.
* Do not leak internal details in error messages.

---

## Handling insufficient context

If required details are missing:

* **Do not invent new APIs/contracts** beyond minimal necessity.
* Add a **brief TODO comment** in the function body indicating the missing info and your safest assumption.
* Keep assumptions conservative and standard for microservices.

---

## Safety / observability defaults

* Validate and sanitize any user input before use.
* Never hardcode secrets, tokens, or environment-specific URLs.
* If a logger/tracer/metrics interface exists in context or on `Server`, use it; otherwise omit observability rather than invent it.

---

## If you add anything new

If you add/extend interfaces, add models, or require new `Server` fields:

* List new `Server` fields in `## added_fields`.
* Put full definitions of **new or extended** interfaces in `## interfaces` (do not repeat unchanged ones).
* Put full definitions of new models in `## added_models`.
* Prefix new models with `models.` inside the handler function body.
  **Do not prefix inside `## added_models`.**
* Describe every interface you output in `## interfaces_description` using the required structured format.

---

## Output format (exactly five sections, in order)

1. `## function` — one Go code block with **only** the body of the function (no signature, no braces).
2. `## added_fields` — each new field as
   `- FieldName FieldType [Optional: Tag, Comment]`
   or `none`.
3. `## interfaces` — one Go code block listing all **new or extended** interfaces (with full definitions), or `none`.
4. `## added_models` — one Go code block listing all new models (full definitions), or `none`.
5. `## interfaces_description` — structured descriptions for interfaces from section 3, or `none`.
   Format:

```
Interface: <Name>
Description: <description of what this interface represents and why it is needed>

Methods:
  - <MethodName>
    Signature: <full Go method signature>
    Description: <what the method does and why it is needed>
```

---



--- Use request context; do not log secrets; validate inputs; prefer structured logging if logger exists in Server. ---