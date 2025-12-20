import json
from playwright.sync_api import sync_playwright

def verify_clear_search_button():
    with sync_playwright() as p:
        browser = p.chromium.launch()
        context = browser.new_context()
        page = context.new_page()

        # Mock user data and token in localStorage
        auth_token = "mock_token"
        user = {
            "id": 1,
            "email": "test@example.com",
            "role": "user"
        }

        # Inject localStorage before navigating
        page.add_init_script(f"""
            localStorage.setItem('auth_token', '{auth_token}');
            localStorage.setItem('user', '{json.dumps(user)}');
        """)

        # Enable console logging
        page.on("console", lambda msg: print(f"Browser console: {msg.text}"))

        # Mock API responses
        # Mock auth check
        page.route("**/api/v1/auth/me", lambda route: route.fulfill(
            status=200,
            content_type="application/json",
            body=json.dumps(user)
        ))

        # Mock monitors response
        page.route("**/api/v1/monitors", lambda route: route.fulfill(
            status=200,
            content_type="application/json",
            body=json.dumps([
                {
                    "id": "1",
                    "name": "Test Monitor 1",
                    "url": "https://example.com",
                    "type": "http",
                    "status": "up",
                    "check_interval": 60,
                    "created_at": "2023-01-01T00:00:00Z"
                },
                {
                    "id": "2",
                    "name": "Test Monitor 2",
                    "url": "https://test.com",
                    "type": "ping",
                    "status": "down",
                    "check_interval": 300,
                    "created_at": "2023-01-02T00:00:00Z"
                }
            ])
        ))

        # Navigate to monitors page
        page.goto("http://localhost:3000/user/monitors")

        # Wait for monitors to load
        page.wait_for_selector("text=Test Monitor 1")

        # Type into search box
        search_input = page.get_by_placeholder("Search monitors...")
        search_input.fill("Test")

        # Wait for clear button to appear (it's conditional on searchQuery)
        clear_button = page.get_by_label("Clear search")
        clear_button.wait_for(state="visible")

        # Take screenshot of search with text and clear button
        page.screenshot(path=".Jules/verification/search_with_text.png")
        print("Screenshot 1 taken: Search input with text and clear button")

        # Click clear button
        clear_button.click()

        # Verify input is cleared
        search_input_value = search_input.input_value()
        assert search_input_value == "", f"Expected empty input, got '{search_input_value}'"

        # Verify clear button is hidden
        # The element might be removed from DOM or hidden.
        # Since it's inside {#if searchQuery}, it should be removed from DOM.
        clear_button.wait_for(state="detached")

        # Take screenshot of cleared search
        page.screenshot(path=".Jules/verification/search_cleared.png")
        print("Screenshot 2 taken: Search input cleared")

        browser.close()

if __name__ == "__main__":
    try:
        verify_clear_search_button()
        print("Verification script completed successfully.")
    except Exception as e:
        print(f"Verification script failed: {e}")
        exit(1)
