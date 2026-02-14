---
name: security-audit-skills
description: Scans the agent's skills folder for malicious patterns, 
             prompt injections, and unauthorized data exfiltration risks.
---

# Instructions
1. Locate the skills directory (`.agents/skills`).
2. Run `make skills-scan` to scan all ./skills recursively.
3. Flag any skill with a 'High' severity rating for manual review.
4. Provide a structured report of any suspicious YARA matches or insecure instructions found in SKILL.md files.