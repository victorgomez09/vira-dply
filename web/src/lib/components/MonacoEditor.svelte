<script lang="ts">
	import { onMount, onDestroy } from 'svelte';
	import type * as Monaco from 'monaco-editor';

	interface Props {
		value?: string;
		language?: string;
		theme?: string;
		height?: string;
		readonly?: boolean;
		onchange?: (value: string) => void;
	}

	let {
		value = $bindable(''),
		language = 'sql',
		theme = 'vs',
		height = '400px',
		readonly = false,
		onchange
	}: Props = $props();

	let editorContainer: HTMLDivElement;
	let editor: Monaco.editor.IStandaloneCodeEditor | undefined;
	let monaco: typeof Monaco | undefined;

	onMount(async () => {
		monaco = await import('monaco-editor');
		
		editor = monaco.editor.create(editorContainer, {
			value,
			language,
			theme,
			readOnly: readonly,
			minimap: { enabled: false },
			fontSize: 14,
			lineNumbers: 'on',
			roundedSelection: false,
			scrollBeyondLastLine: false,
			automaticLayout: true,
			tabSize: 2
		});

		editor.onDidChangeModelContent(() => {
			const newValue = editor!.getValue();
			value = newValue;
			if (onchange) {
				onchange(newValue);
			}
		});
	});

	onDestroy(() => {
		if (editor) {
			editor.dispose();
		}
	});

	$effect(() => {
		if (editor && value !== editor.getValue()) {
			editor.setValue(value);
		}
	});
</script>

<div bind:this={editorContainer} style="height: {height}; width: 100%;"></div>
