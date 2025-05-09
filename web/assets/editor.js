// wrap in IIFE to avoid polluting globals
function loadMonacoCustom(){
	// tell require.js where to find Monaco
	require.config({
		paths: { vs: 'https://cdn.jsdelivr.net/npm/monaco-editor@0.47.0/min/vs' }
	});

	// the snippets for each language
	const defaultSnippets = {
		go: `package main

func main() {
    // your Go solution here
}
`,
		python: `def main():
    # your Python solution here

if __name__ == "__main__":
    main()
`,
		javascript: `function main() {
    // your JS solution here
}

main();
`,
		cpp: `int main() {
  // your C++ solution here
  return 0;
}`,
		c: `int main() {
  // your C solution here
  return 0;
}`,

		rust: `fn main() {
    // your Rust solution here
}`
	};

	// create or re-create the editor in #code-editor
	function initMonaco(initialCode) {
		const editorContainer = document.getElementById('code-editor');
		if (!editorContainer) {
			// No editor on this page, don't do anything
			return;
		}
		console.log("Initiating Monaco!")
		require(['vs/editor/editor.main'], function() {
			window.editor = monaco.editor.create(
				document.getElementById('code-editor'),
				{
					value: initialCode || defaultSnippets[ document.getElementById('language-select').value ],
					language: document.getElementById('language-select').value,
					automaticLayout: true
				}
			);
			// wire up Reset button
			document.getElementById('reset-button').addEventListener('click', resetLanguage);
			// wire up language dropdown
			document.getElementById('language-select').addEventListener('change', resetLanguage);
		});
	}

	// change the language and reset to its snippet
	function resetLanguage() {
		const lang = document.getElementById('language-select').value;
		monaco.editor.setModelLanguage(editor.getModel(), lang);
		editor.setValue(defaultSnippets[lang] || '');
	}

	// when the page first loads, bootstrap the editor
	document.addEventListener('DOMContentLoaded', function(){
		initMonaco();
	});

	// after any HTMX swap, if we've swapped in the editor container, re-init
	document.addEventListener('htmx:afterSwap', function(evt){
		// adjust the selector to what you swap into
		if (evt.detail.target.id === 'main') {
			// assume no server-provided initial code for simplicity
			console.log("Reseting Monaco!")
			initMonaco();
		}
		if (evt.detail.target.id === 'results') {
			evt.detail.target.scrollTop = 0;
		}
	});

};

loadMonacoCustom();
