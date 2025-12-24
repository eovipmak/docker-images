## 2025-05-26 - Search Input Clear Button
**Learning:** Users on mobile devices often struggle to clear long search queries. Adding a visible "Clear" button within the input field (revealed when text is present) significantly improves usability.
**Action:** Implement the "Search with Clear" pattern for all text filters: use `pr-10` on the input and position an absolute clear button on the right.

## 2025-05-27 - Monitor Modal Autofocus
**Learning:** When users open a creation modal (like "Add Monitor"), their primary intent is to start typing immediately. Requiring an extra click to focus the name field creates unnecessary friction.
**Action:** Always add the `autofocus` attribute to the first input field in creation/edit modals to support immediate keyboard interaction.
