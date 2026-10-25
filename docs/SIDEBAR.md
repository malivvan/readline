* [Examples](../cmd/README.md)
* [Home](./HOME.md)
------
## Examples

<details>
<summary><a href="../cmd/README.md">Runnable examples</a></summary>
<ul>
        <li><a href="../cmd/basics">basics - the Readline loop and its errors</a></li>
        <li><a href="../cmd/completions">completions - values, tags, styles and live refresh</a></li>
        <li><a href="../cmd/history">history - file, memory and multiple sources</a></li>
        <li><a href="../cmd/keymaps">keymaps - custom commands, binds and modes</a></li>
        <li><a href="../cmd/highlighting">highlighting - the SyntaxHighlighter callback</a></li>
        <li><a href="../cmd/multiline">multiline - the AcceptMultiline callback</a></li>
        <li><a href="../cmd/prompts">prompts - all prompt handlers and logging</a></li>
        <li><a href="../cmd/suggestions">suggestions - inline ghost text</a></li>
        <li><a href="../cmd/paste">paste - bracketed paste transformation</a></li>
        <li><a href="../cmd/options">options - programmatic configuration</a></li>
        <li><a href="../cmd/macros">macros - macro binds and replay</a></li>
    </ul>
</details>

----

### Users

<details>
<summary><a href="./OPTIONS.md">Options</a></summary>
<ul>
        <li><a href="./OPTIONS.md#introduction">Introduction</a></li>
        <li><a href="./OPTIONS.md#setting-options-in-the-inputrc-file">Setting options in the <code>.inputrc</code> file</a></li>
        <li><a href="./OPTIONS.md#options">Options</a></li>
    </ul></li>
</ul>
</details>

<details>
<summary><a href="./KEYMAPS-&-COMMANDS.md">Keymaps & Commands</a></summary>
<ul>
    <li><a href="#table-of-contents">Table of Contents</a></li>
    <li><a href="./KEYMAPS-&-COMMANDS.md#keymaps">Keymaps</a><ul>
        <li><a href="./KEYMAPS-&-COMMANDS.md#available-keymaps">Available keymaps</a></li>
        <li><a href="./KEYMAPS-&-COMMANDS.md#key-dispatch--matching">Key dispatch & matching</a></li>
        <li><a href="./KEYMAPS-&-COMMANDS.md#example">Example</a></li>
    </ul></li>
    <li><a href="./KEYMAPS-&-COMMANDS.md#commands">Commands</a><ul>
        <li><a href="./KEYMAPS-&-COMMANDS.md#movement">Movement</a></li>
        <li><a href="./KEYMAPS-&-COMMANDS.md#modifying-text">Modifying Text</a></li>
        <li><a href="./KEYMAPS-&-COMMANDS.md#arguments">Arguments</a></li>
        <li><a href="./KEYMAPS-&-COMMANDS.md#macros">Macros</a></li>
        <li><a href="./KEYMAPS-&-COMMANDS.md#history-control">History Control</a></li>
        <li><a href="./KEYMAPS-&-COMMANDS.md#completion">Completion</a></li>
        <li><a href="./KEYMAPS-&-COMMANDS.md#miscellaneous">Miscellaneous</a></li>
        <li><a href="./KEYMAPS-&-COMMANDS.md#text-objects">Text Objects</a></li>
    </ul></li>
</ul>
</details>

<details>
<summary><a href="./DEFAULT-BINDS-&-OPTIONS.md">Default Binds & Options</a></summary>
<ul>
    <li><a href="./DEFAULT-BINDS-&-OPTIONS.md#options">Options</a></li>
    <li><a href="./DEFAULT-BINDS-&-OPTIONS.md#emacs">Emacs</a><ul>
        <li><a href="./DEFAULT-BINDS-&-OPTIONS.md#emacs-standard">Emacs standard</a></li>
        <li><a href="./DEFAULT-BINDS-&-OPTIONS.md#emacs-meta">Emacs Meta</a></li>
        <li><a href="./DEFAULT-BINDS-&-OPTIONS.md#emacs-ctrlx">Emacs CtrlX</a></li>
    </ul></li>
    <li><a href="./DEFAULT-BINDS-&-OPTIONS.md#vim">Vim</a><ul>
        <li><a href="./DEFAULT-BINDS-&-OPTIONS.md#insert">Insert</a></li>
        <li><a href="./DEFAULT-BINDS-&-OPTIONS.md#command">Command</a></li>
        <li><a href="./DEFAULT-BINDS-&-OPTIONS.md#visual">Visual</a></li>
        <li><a href="./DEFAULT-BINDS-&-OPTIONS.md#operator-pending">Operator Pending</a></li>
    </ul></li>
    <li><a href="./DEFAULT-BINDS-&-OPTIONS.md#completion--isearch">Completion & Isearch</a><ul>
        <li><a href="./DEFAULT-BINDS-&-OPTIONS.md#menu-select">Menu-Select</a></li>
        <li><a href="./DEFAULT-BINDS-&-OPTIONS.md#incremental-search">Incremental Search</a></li>
        <li><a href="./DEFAULT-BINDS-&-OPTIONS.md#non-incremental-search">Non-incremental Search</a></li>
    </ul></li></ul></li>
</ul>
</details>

<details>
<summary><a href="./MACROS.md">Macros</a></summary>
<ul>
        <li><a href="./MACROS.md#emacs-mode">Emacs mode</a></li>
        <li><a href="./MACROS.md#vim-mode">Vim mode</a></li>
    </ul></li></ul></li>
</ul>
</details>

<details>
<summary><a href="./VIM-MODE.md">Vim (additional features)</a></summary>
<ul>
        <li><a href="./VIM-MODE.md#registers">Registers</a></li>
        <li><a href="./VIM-MODE.md#text-objects-examples">Text objects (examples)</a></li>
    </ul></li>
</ul>
</details>


----
### Developers 

<details>
<summary><a href="./INTRODUCTION.md">Introduction</a></summary>
<ul>
        <li><a href="./INTRODUCTION.md#current-shell-libraries">Current shell libraries</a></li>
        <li><a href="./INTRODUCTION.md#original-readline-compatiblity">Original readline compatiblity</a></li>
        <li><a href="./INTRODUCTION.md#aims-and-choices-of-this-package">Aims and choices of this package</a><ul>
            <li><a href="./INTRODUCTION.md#architecture--inspirations">Architecture & inspirations</a></li>
            <li><a href="./INTRODUCTION.md#simpler-apis-same-user-experience">Simpler APIs, same user experience</a></li>
        </ul></li></ul></li>
</ul>
</details>

<details>
<summary><a href="./CONFIGURATION-FILE.md">Configuration File</a></summary>
<ul>
            <li><a href="./CONFIGURATION-FILE.md#loading-configurations">Loading configurations</a></li>
            <li><a href="./CONFIGURATION-FILE.md#global-options">Global options</a></li>
            <li><a href="./CONFIGURATION-FILE.md#key-bindings">Key bindings</a></li>
        </ul></li></ul></li>
</ul>
</details>

<details>
<summary><a href="./HISTORY-SOURCES.md">History Sources</a></summary>
<ul>
        <li><a href="./HISTORY-SOURCES.md#file-based-command-history">File-based command history</a></li>
        <li><a href="./HISTORY-SOURCES.md#adding-and-removing-histories">Adding and removing histories</a></li>
        <li><a href="./HISTORY-SOURCES.md#implementing-history-sources">Custom history sources</a></li>
    </ul></li>
</ul>
</details>

<details>
<summary><a href="./PROMPTS.md">Prompts</a></summary>
<ul>
        <li><a href="./PROMPTS.md#prompt-handlers">Prompt handlers</a></li>
        <li><a href="./PROMPTS.md#logging-to-the-shell">Logging to the shell</a><ul>
            <li><a href="./PROMPTS.md#transient-log">Transient log</a></li>
            <li><a href="./PROMPTS.md#classic-prompt-log">Classic prompt log</a></li>
            <li><a href="./PROMPTS.md#demo">Demo</a></li>
        </ul></li></ul></li>
</ul>
</details>

<details>
<summary><a href="./PROGRAMMABLE-COMMANDS.md">Programmable Commands</a></summary>
<ul>
        <li><a href="./PROGRAMMABLE-COMMANDS.md#introduction">Introduction</a></li>
        <li><a href="./PROGRAMMABLE-COMMANDS.md#accessing-readline-components">Accessing readline components</a></li>
        <li><a href="./PROGRAMMABLE-COMMANDS.md#the-keymap-engine">The keymap engine</a></li>
        <li><a href="./PROGRAMMABLE-COMMANDS.md#example-1-adding-a-url-selection-command">Example 1: Adding a URL-selection command</a></li>
        <li><a href="./PROGRAMMABLE-COMMANDS.md#example-2-adding-a-surround-change-command">Example 2: Adding a surround-change command</a></li>
        <li><a href="./PROGRAMMABLE-COMMANDS.md#key-event-callbacks">Key-event callbacks</a></li>
    </ul></li>
</ul>
</details>

<details>
<summary><a href="./COMPLETIONS-&-HINTS.md">Completions & Hints</a></summary>
<ul>
        <li><a href="./COMPLETIONS-&-HINTS.md#the-completions-container">The <code>Completions</code> container</a></li>
        <li><a href="./COMPLETIONS-&-HINTS.md#completion-candidates">Completion Candidates</a></li>
        <li><a href="./COMPLETIONS-&-HINTS.md#autosuffix-removal">Autosuffix removal</a></li>
        <li><a href="./COMPLETIONS-&-HINTS.md#hints-and-usage-strings">Hints and usage strings</a></li>
        <li><a href="./COMPLETIONS-&-HINTS.md#the-hint-area-and-its-lanes">The hint area and its lanes</a></li>
        <li><a href="./COMPLETIONS-&-HINTS.md#asynchronous-updates">Asynchronous updates</a></li>
    </ul></li>
</ul>
</details>


<details>
<summary><a href="./SYNTAX-HIGHLIGHTING.md">Syntax Highlighting</a></summary>
<ul>
        <li><a href="./SYNTAX-HIGHLIGHTING.md#builtin-highlighting">Builtin highlighting</a></li>
        <li><a href="./SYNTAX-HIGHLIGHTING.md#writing-the-callback">Writing the callback</a></li>
    </ul></li>
</ul>
</details>

<details>
<summary><a href="./INLINE-SUGGESTIONS.md">Inline Suggestions</a></summary>
<ul>
        <li><a href="./INLINE-SUGGESTIONS.md#the-api">The API</a></li>
        <li><a href="./INLINE-SUGGESTIONS.md#setting-a-suggestion">Setting a suggestion</a></li>
        <li><a href="./INLINE-SUGGESTIONS.md#accepting-a-suggestion">Accepting a suggestion</a></li>
    </ul></li>
</ul>
</details>

<details>
<summary><a href="./BRACKETED-PASTE.md">Bracketed Paste</a></summary>
<ul>
        <li><a href="./BRACKETED-PASTE.md#the-paste-transformer">The paste transformer</a></li>
        <li><a href="./BRACKETED-PASTE.md#example">Example</a></li>
    </ul></li>
</ul>
</details>

<details>
<summary><a href="./MULTILINE.md">Multiline support</a></summary>
<ul>
        <li><a href="./MULTILINE.md#writing-the-callback">Writing the callback</a></li>
        <li><a href="./MULTILINE.md#example">Example</a></li>
    </ul></li>
</ul>
</details>




<!-- ---- -->
