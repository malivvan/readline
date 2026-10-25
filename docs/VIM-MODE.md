> **Runnable example:** [`cmd/keymaps`](../cmd/keymaps)

# Table of Contents

<details>
<summary>Click to expand Table of Contents</summary>
<ul>
        <li><a href="#table-of-contents">Table of Contents</a></li>
        <li><a href="#registers">Registers</a></li>
        <li><a href="#text-objects-examples">Text objects (examples)</a></li>
    </ul></li>
</ul>
</details>

This readline library probably offers the most complete Vim edition mode out there.
If you want to configure keybinds for each keymap, see the [configuration](./CONFIGURATION-FILE.md) and the [commands](./KEYMAPS-&-COMMANDS.md).

This page gives additional explanations and details on features found in Vim modes.
For Vim-style macro recording and execution, see the [Macros page](./MACROS.md).

## Registers

Registers are implemented in this library. The following registers are usable:

* The default `""` register, which is used for many actions (`yank`, `cut`, etc)
* 10 numbered registers, to which buffers are automatically added (`"0`, or `"5`, etc)
* 26 lettered registers (lowercase), to which you can append with `"D` (D being the uppercase of the `"d` register)
* Triggered in Insert Mode with `Alt"` (buggy sometimes: goes back to Normal mode selecting a register, will have to fix this)

**Example Usage**

Yank/paste operations of any sort can occur and be assigned to registers. 
An example sequence that should be familiar to Vim users:

* To copy to the `d` register the next 4 words: `"d y4w`
* To append to this `d` register the cuttend end of line: `"D d$`
* In this example, the `d` register buffer is also the buffer in the default register `""`
* You could either:
    - Paste 3 times this buffer while in Normal mode: `3p`
    - Paste the buffer once in Insert mode: `CtrlY`

You can also show and cycle through your current registers (as completions) with `Alt-Shift-R"` by default.


## Text objects (examples)
