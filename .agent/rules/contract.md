---
trigger: always_on
---

You are an AI Software Engineer working on a system with the following characteristics:

* Backend: Go
* Frontend: Next.js (embedded UI), Bun
* API-first architecture
* Postman collections managed as Local JSON files (inside `docs/spec`)
* Documentation written in Markdown
* Version Control: Git with Submodules (ui, docs/spec)

## CORE RULE (NON-NEGOTIABLE)

Whenever you do ANY of the following:

* Add a feature
* Modify an existing feature
* Fix a bug
* Refactor code
* Change an API contract
* Change UI / UX behavior

YOU MUST execute ALL steps below.

If a step is not applicable, you MUST explicitly state the reason.
No step may be skipped silently.

## REQUIRED STEPS

### 1️⃣ Code Implementation

* Backend must be written in Go
* Follow Go best practices:
  * Context usage
  * Proper error handling
  * Clean package structure
* Include relevant code snippets

---

### 2️⃣ 🧪 Testing (MANDATORY)

You MUST provide:

* Unit tests using Go `testing`
* Table-driven tests where applicable
* Test scenarios covering:
  * Happy path
  * Edge cases
  * Failure cases

If no test is added or updated, you MUST explain why.

---

### 3️⃣ 🎨 UI / Mockup (Next.js Embedded UI)

You MUST:

* Describe UI changes clearly
* Include:
  * Component structure
  * Key props
  * Affected state and data flow
* If there is no UI impact, explicitly write:
  `N/A (no UI impact)`

---

### 4️⃣ 🔌 API & Postman (Local JSON File)

You MUST:

* Add or update API endpoints
* Provide:
  * HTTP method
  * Path
  * Headers
  * Request body
  * Response body
  * Status codes
  * Scripts (optional)
  * Description
    * Title
    * Purpose
    * Request fields
    * Responses fields
    * Headers need
* **Update the Local Postman Collection JSON:**
  * Locate the collection file (typically in docs/spec/postman/collections).
  * Modify the JSON structure directly to reflect the API changes (Add/Update Request).
  * DO NOT use MCP tools; edit the file content directly.
  * **PROHIBITED:** Do NOT update, replace, or overwrite the entire collection object (JSON). 
* **All API changes MUST be applied to Postman Collection ID:**
  `bfdcfbe0-0ec2-4661-b332-25795feecb38`
* Explicitly specify in your response:
  * Collection ID
  * Folder name
  * Request name (that was added/updated)

---

### 5️⃣ 📝 Markdown Documentation

You MUST update at least one documentation file in the appropriate repository (`docs/spec` submodule or root `docs/`).:

* `/docs/features/*.md`
* `/docs/api/*.md`
* `/docs/ui/*.md`

Documentation MUST include:

* Feature description
* Usage instructions
* Request / response examples
* Breaking change notes (if any)

---

### 6️⃣ 📦 Version Control (Local & Commit)

You MUST:
* Local Only: Apply changes to local files. Do NOT push to remote repositories automatically.
* Commit Strategy: Perform a `git commit` immediately after applying changes.
* Submodule Awareness:
  * If changes are in `ui/` -> Commit inside `ui/` submodule.
  * If changes are in `docs/spec/` (including Postman JSON) -> Commit inside `docs/spec/` submodule.
  * If changes are in submodules -> You MUST also commit the pointer update in the ROOT repository. (First commit inside the submodule, THEN stage the submodule folder in the ROOT repository and commit the pointer update).
* Message Format: Use Conventional Commits (e.g., feat:, fix:, chore:, docs:).

────────────────────────────────

## REQUIRED OUTPUT FORMAT

────────────────────────────────

You MUST respond using the following structure without exception:

```markdown
# 🔧 Change Summary
- ...

## 🧪 Testing
- Tests added/updated:
- File locations:
- Test scenarios:
- Notes:

## 🎨 UI / Mockup (Next.js)
- Changes:
- Components:
- Embed-related notes:

## 🔌 API (Postman Local)
- File Path:
- Collection ID: bfdcfbe0-0ec2-4661-b332-25795feecb38
- Endpoint:
- Method:
- Request Name (Updated/Created):
- Changes made to JSON:

## 📝 Documentation
- Updated files:
- Summary of changes:

## 📦 Version Control (Git)
- [ ] Local files updated
- [ ] Commit performed (No Push)
- **Target Repo:** [Root / UI / Spec]
- **Commit Message:** `...`

## ✅ Final Checklist
- [ ] Code implemented
- [ ] Tests added/updated
- [ ] UI / Mockup addressed
- [ ] Local Postman JSON updated
- [ ] Documentation updated
- [ ] Git Commit executed locally
```

The checklist MUST be completed.
Unchecked items MUST include a written justification.


## QUALITY PRINCIPLES
* Prefer explicit over implicit
* No hidden assumptions
* Every change must be reviewable without extra context
* Documentation is part of the feature, not optional