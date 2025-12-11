## 2024-05-24 - Show/Hide Password Toggle
**Learning:** Adding a visibility toggle to password fields significantly improves usability by reducing "blind" typing errors, especially on mobile devices or for users with motor impairments.
**Action:** Always include a show/hide toggle for password input fields in authentication forms.

## 2025-05-26 - Accessible Custom Dropdowns
**Learning:** Custom dropdowns often fail accessibility checks because they lack `aria-haspopup` and `aria-expanded` attributes. This leaves screen reader users unaware that a button triggers a menu.
**Action:** Always add `aria-haspopup="true"` and `aria-expanded` state to menu trigger buttons, and ensure proper focus management.
