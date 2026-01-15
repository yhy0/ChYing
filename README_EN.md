# ChYing: Three Years in the Making, A Security Professional's Open Source Dream

<p align="center">
  <img src="./images/b5e9f48a-d10c-4428-bad9-1673f1084af8.png" width="400" alt="ChYing">
</p>

> Preface
>
> In April 2023, I created a project called "ChYing" on GitHub.
>
> ChYing (承影) is named after an ancient legendary sword. According to "Liezi - Chapter of Tang Wen": "The second is ChengYing (Shadow Bearer). At the transition between dawn and daybreak, or between dusk and darkness, if you look northward, you can faintly see something exists, though its form is indiscernible. When it touches something, there is a subtle sound, yet objects it passes through remain unharmed." It refers to a divine sword that can only be glimpsed in the interplay of light and shadow.
>
> I hope this tool can be the same — becoming a sharp weapon in the hands of security professionals, operating in the shadows of penetration testing.

## Origin: A Security Professional's Product Dream

I am a cybersecurity enthusiast who loves developing practical and interesting tools.

I've always had a dream: to create a security product as widely used as Xray or BurpSuite.

Before ChYing, I created another open source project — [Jie](https://github.com/yhy0/Jie), a comprehensive web security assessment tool supporting active/passive scanning, vulnerability detection, and information gathering. It has earned 600+ stars so far.

But Jie is more oriented towards automated scanning — the kind of tool you "run and forget." I still needed an interactive penetration testing platform — one that can capture packets, modify them, replay them, and brute force, like Burp Suite, but lighter, more modern, and more "mine."

Thus, ChYing was born.

Moreover, I integrated Jie's scanning capabilities into it. ChYing + Jie: one handles interaction, the other handles automation. They complement each other.

## Three Years: Countless Late Nights

From 2023 to now, this project has been through a lot:

**First Version (2023):** Built a framework with Wails + Vue, implemented basic proxy, directory scanning, and JWT parsing. The interface was crude, but functional. After open-sourcing, it gained 400+ stars — honestly, I was a bit surprised.

**Dormant Period (2023-2024):** Work got busy, and the project went on hiatus. But it was always on my mind; I felt something was still missing.

**Refactoring Period (2024-2025):** Created a private repository, 137 commits, countless late nights, and a major overhaul.

Here's a turning point worth mentioning: the emergence of LLMs.

Honestly, my frontend skills are quite limited, and I didn't have time to study systematically. The previous interface was crude because that was the best I could do.

But after AI programming tools like Cursor came out, everything changed. I no longer needed to struggle with "how to implement this animation effect" or "should I use flex or grid for this layout" — I just needed to describe what I wanted, and AI helped me implement it.

**AI is an amplifier of personal capability.** It won't think about how to design your product for you, but it can help turn your ideas into code. For someone like me who "has ideas but technical limitations," this was the key to breaking through.

Finally, the "modern UI" I had envisioned for so long could become reality.

**Now: This is what it looks like**

![image-20260112204045831](./images/image-20260112204045831.png)

![image-20260112204151347](./images/image-20260112204151347.png)

## Core Features

**HTTP Proxy & Traffic Analysis**
- Real-time HTTP/HTTPS traffic capture
- Smart filtering (by method, host, status code, path)
- Right-click menu to send to Repeater/Intruder/Scanner

**Repeater**
- Manually modify requests, test repeatedly
- Multi-tab support for easier comparison testing

**Intruder**
- Automated attack testing
- Multiple payload types supported
- Real-time result display

**Decoder**
- One-click URL/Base64/Hex/Unicode conversion
- MD5/SHA hash calculation
- Chain encoding/decoding support

**Plugin Modules**
- JWT parsing and key brute-forcing
- Swagger API testing (unauthorized access, injection detection)
- 403 Bypass
- Shiro decryption

**Integrated Jie Scanner**
- Passive traffic scanning
- Active vulnerability detection (XSS, SQL injection, SSRF, command execution, etc.)
- Nuclei POC support

## Tech Stack

- **Backend:** Go (high performance, cross-platform)
- **Frontend:** Vue 3 + TypeScript
- **Framework:** Wails v3 (perfect combination of Go + Web)
- **Database:** SQLite (lightweight local storage)
- **Scanning Engine:** Jie

## UI: Liquid Glass Design

The biggest change in this refactoring is the UI. With AI's help, I finally achieved the design I had in mind:

- **Glassmorphism style:** Semi-transparent, blurred backgrounds, soft shadows
- **Dark/Light themes:** Easy on the eyes and aesthetically pleasing
- **Responsive layout:** Works on various resolutions
- **Smooth animations:** Every interaction has feedback

I'm not a professional designer, but I hope that every time I open this tool, it brings a pleasant feeling.

## Why Open Source Now?

This project has been sitting in my private repository for almost two years.

I kept not open-sourcing it, always feeling "it's not good enough" — features incomplete, bugs unfixed, UI could be better... Plus, my work focus shifted, I did less penetration testing, and hadn't opened Burp in a long time.

But recently I realized: **perfection is the enemy of open source.**

Rather than letting it continue to sit on my hard drive, better to release it. Maybe someone will find it useful, maybe someone will help improve it, maybe it will inspire others to do similar things.

More importantly, I want to leave a trace of these three years of effort. Code becomes outdated, technology iterates, but I hope this intention of "wanting to create something" will be remembered.

## To Fellow Security Professionals

If you're also in the security field, I want to say:

This tool is not meant to replace Burp Suite. Burp remains the most professional and comprehensive penetration testing tool.

But if you:
- Want a lightweight tool that starts faster
- Want a more modern interface option
- Want an open source solution you can modify yourself
- Or just want to see how a security professional tinkers with their own tools

Then ChYing might be worth a try.

## Finally

Three years, 137 commits, countless late nights.

From a crude prototype to what it is now.

It's not perfect, there's still much to improve. But it's mine, typed out line by line.

Now, it can be yours too.

---
ChYing GitHub: https://github.com/yhy0/ChYing

Jie GitHub: https://github.com/yhy0/Jie

**Give it a Star?** That's the best encouragement for an independent developer.

---
ChYing — May you find your own edge in the interplay of light and shadow.

---
