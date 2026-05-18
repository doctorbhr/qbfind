#!/usr/bin/env python3
"""
GitHub Issue Scanner - Open Source Contributor Bot
Designed for scanning "good first issue" tasks across public GitHub repositories.
Works out of the box with ZERO dependencies (pure Python standard library).

Maintainer: tommyvercetti89
AI Partner: Antigravity (Google DeepMind)
"""

import json
import urllib.request
import urllib.parse
import sys
import time

# Premium ANSI Terminal Colors
C_BLUE = "\033[94m"
C_GREEN = "\033[92m"
C_YELLOW = "\033[93m"
C_CYAN = "\033[96m"
C_RED = "\033[91m"
C_MAGENTA = "\033[95m"
C_BOLD = "\033[1m"
C_DIM = "\033[2m"
C_RESET = "\033[0m"

def display_banner():
    banner = f"""{C_MAGENTA}{C_BOLD}
   ┌────────────────────────────────────────────────────────┐
   │             🤖 GITHUB ISSUE SCANNER BOT                │
   │      Scan "Good First Issues" & Code with AI           │
   └────────────────────────────────────────────────────────┘{C_RESET}"""
    print(banner)

def scan_github_issues(language="go", label="good first issue", token=None):
    """
    Scans public GitHub repositories for open issues matching a specific language and label.
    """
    print(f"\n{C_CYAN}🔄 Scanning GitHub for issues...{C_RESET}")
    print(f"   • Language: {C_BOLD}{language.upper()}{C_RESET}")
    print(f"   • Label:    {C_BOLD}{label}{C_RESET}")

    # Build the GitHub Search API query
    # Example: is:issue is:open label:"good first issue" language:go
    query_str = f'is:issue is:open label:"{label}" language:{language}'
    encoded_query = urllib.parse.quote(query_str)
    url = f"https://api.github.com/search/issues?q={encoded_query}&sort=created&order=desc&per_page=15"

    headers = {
        "Accept": "application/vnd.github.v3+json",
        "User-Agent": "GitHub-Issue-Scanner-Bot-TommyVercetti"
    }

    # Add GitHub Personal Access Token (PAT) if provided to bypass rate limits
    if token:
        headers["Authorization"] = f"token {token}"
        print(f"   • Auth:     {C_GREEN}GitHub Token Enabled (Higher Rate Limit){C_RESET}")
    else:
        print(f"   • Auth:     {C_YELLOW}Anonymous (Limit: 60 req/hour. Provide a PAT if needed){C_RESET}")

    req = urllib.request.Request(url, headers=headers)

    try:
        with urllib.request.urlopen(req) as response:
            data = json.loads(response.read().decode())
            items = data.get("items", [])
            
            if not items:
                print(f"\n{C_YELLOW}⚠️ No issues found matching those criteria. Try another language or label!{C_RESET}")
                return

            print(f"\n{C_GREEN}✔ Found {len(items)} recent open issues! Rendering table...{C_RESET}\n")
            print(f"{C_BOLD}{'=' * 95}{C_RESET}")
            print(f"{C_BOLD}{C_BLUE}{'Repository':<28} | {'Title':<45} | {'Created At'}{C_RESET}")
            print(f"{C_BOLD}{'=' * 95}{C_RESET}")

            for item in items:
                # Extract repo name from html_url (e.g. https://github.com/user/repo/issues/1)
                html_url = item.get("html_url", "")
                parts = html_url.split("/")
                repo_display = f"{parts[3]}/{parts[4]}" if len(parts) > 4 else "Unknown"

                title = item.get("title", "")
                # Truncate title if too long to maintain clean grid layout
                if len(title) > 42:
                    title = title[:39] + "..."

                created_at = item.get("created_at", "")[:10]  # Get YYYY-MM-DD
                
                print(f"{C_YELLOW}{repo_display:<28}{C_RESET} | {title:<45} | {C_DIM}{created_at}{C_RESET}")
                print(f"   {C_CYAN}🔗 URL:{C_RESET} {C_GREEN}{html_url}{C_RESET}")
                
                # Render labels if present
                labels = [l.get("name") for l in item.get("labels", [])]
                if labels:
                    labels_str = ", ".join(labels[:4])
                    print(f"   {C_DIM}🏷️ Labels: {labels_str}{C_RESET}")
                print(f"{C_DIM}{'-' * 95}{C_RESET}")

    except urllib.error.HTTPError as e:
        print(f"\n{C_RED}❌ HTTP Error {e.code}: {e.reason}{C_RESET}")
        if e.code == 403:
            print(f"{C_RED}   Rate limit exceeded or forbidden. Please wait a few minutes or supply a GitHub Personal Access Token (PAT).{C_RESET}")
    except Exception as e:
        print(f"\n{C_RED}❌ Error occurred: {e}{C_RESET}")

def main():
    display_banner()
    
    # Default search options
    language = "go"
    label = "good first issue"
    token = None  # Add your GitHub PAT string here if you get rate limited

    # Simple interactive CLI options
    print(f"{C_BOLD}Choose search configuration or press ENTER for defaults:{C_RESET}")
    user_lang = input(f"• Target Programming Language (default: {language}): ").strip()
    if user_lang:
        language = user_lang

    user_label = input(f"• Target Label (default: '{label}'): ").strip()
    if user_label:
        label = user_label

    user_token = input(f"• GitHub Personal Access Token [Optional] (press Enter to skip): ").strip()
    if user_token:
        token = user_token

    scan_github_issues(language, label, token)
    
    print(f"\n{C_MAGENTA}{C_BOLD}🤖 Tip:{C_RESET} Feed the Issue Description & File Contents from any of these URLs to your AI model to generate the perfect Pull Request! Good luck!")

if __name__ == "__main__":
    try:
        main()
    except KeyboardInterrupt:
        print(f"\n\n{C_YELLOW}👋 Exiting Scanner. Keep building!{C_RESET}")
        sys.exit(0)
