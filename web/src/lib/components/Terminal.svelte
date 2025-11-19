<script lang="ts">
	import { onMount, onDestroy } from 'svelte';
	import { Terminal } from '@xterm/xterm';
	import { FitAddon } from '@xterm/addon-fit';
	import { WebLinksAddon } from '@xterm/addon-web-links';
	import '@xterm/xterm/css/xterm.css';

	interface Props {
		containerID: string;
		projectID: string;
		endpoint: string;
		onConnect?: () => void;
		onDisconnect?: () => void;
		onError?: (error: Error) => void;
	}

	let {
		containerID,
		projectID,
		endpoint,
		onConnect,
		onDisconnect,
		onError
	}: Props = $props();

	let terminalContainer: HTMLDivElement;
	let terminal: Terminal | null = null;
	let ws: WebSocket | null = null;
	let fitAddon: FitAddon | null = null;

	function connect() {
		if (!terminal) return;

		const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
		const token = localStorage.getItem('auth_token');
		const wsUrl = `${protocol}//${window.location.host}${endpoint}${token ? `?token=${token}` : ''}`;

		try {
			ws = new WebSocket(wsUrl);

			ws.onopen = () => {
				onConnect?.();
				if (fitAddon) {
					const size = fitAddon.proposeDimensions();
					if (size) {
						ws?.send(
							JSON.stringify({
								type: 'resize',
								rows: size.rows,
								cols: size.cols
							})
						);
					}
				}
			};

		ws.onmessage = (event) => {
			try {
				const msg = JSON.parse(event.data);
				if (msg.type === 'output' && msg.data && terminal) {
					terminal.write(msg.data);
				}
			} catch (e) {
				console.error('Failed to parse message:', e);
			}
		};

			ws.onerror = (event) => {
				const error = new Error('WebSocket error');
				onError?.(error);
			};

			ws.onclose = () => {
				onDisconnect?.();
			};
		} catch (e) {
			onError?.(e as Error);
		}
	}

	function disconnect() {
		if (ws) {
			ws.close();
			ws = null;
		}
	}

	onMount(() => {
		terminal = new Terminal({
			cursorBlink: true,
			fontSize: 14,
			fontFamily: 'Menlo, Monaco, "Courier New", monospace',
			theme: {
				background: '#1e1e1e',
				foreground: '#d4d4d4',
				cursor: '#d4d4d4',
				black: '#000000',
				red: '#cd3131',
				green: '#0dbc79',
				yellow: '#e5e510',
				blue: '#2472c8',
				magenta: '#bc3fbc',
				cyan: '#11a8cd',
				white: '#e5e5e5',
				brightBlack: '#666666',
				brightRed: '#f14c4c',
				brightGreen: '#23d18b',
				brightYellow: '#f5f543',
				brightBlue: '#3b8eea',
				brightMagenta: '#d670d6',
				brightCyan: '#29b8db',
				brightWhite: '#e5e5e5'
			}
		});

		fitAddon = new FitAddon();
		terminal.loadAddon(fitAddon);
		terminal.loadAddon(new WebLinksAddon());

		terminal.open(terminalContainer);
		fitAddon.fit();

		terminal.onData((data) => {
			if (ws && ws.readyState === WebSocket.OPEN) {
				ws.send(
					JSON.stringify({
						type: 'input',
						data
					})
				);
			}
		});

		terminal.onResize(({ rows, cols }) => {
			if (ws && ws.readyState === WebSocket.OPEN) {
				ws.send(
					JSON.stringify({
						type: 'resize',
						rows,
						cols
					})
				);
			}
		});

		const resizeObserver = new ResizeObserver(() => {
			if (fitAddon) {
				fitAddon.fit();
			}
		});
		resizeObserver.observe(terminalContainer);

		connect();

		return () => {
			resizeObserver.disconnect();
		};
	});

	onDestroy(() => {
		disconnect();
		if (terminal) {
			terminal.dispose();
		}
	});
</script>

<div class="terminal-wrapper">
	<div bind:this={terminalContainer} class="terminal-container"></div>
</div>

<style>
	.terminal-wrapper {
		width: 100%;
		height: 100%;
		background-color: #1e1e1e;
		border-radius: 4px;
		overflow: hidden;
		display: flex;
		flex-direction: column;
	}

	.terminal-container {
		flex: 1;
		padding: 8px;
		overflow: hidden;
	}
</style>
