export function generatePassword(length: number = 16): string {
	const charset = 'abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789!@#$%^&*';
	let password = '';
	const values = new Uint32Array(length);
	crypto.getRandomValues(values);
	for (let i = 0; i < length; i++) {
		password += charset[values[i] % charset.length];
	}
	return password;
}

export function generateUsername(): string {
	const adjectives = ['smart', 'clever', 'swift', 'bright', 'quick', 'sharp'];
	const nouns = ['user', 'admin', 'dev', 'app', 'db', 'cloud'];
	const adjective = adjectives[Math.floor(Math.random() * adjectives.length)];
	const noun = nouns[Math.floor(Math.random() * nouns.length)];
	const number = Math.floor(Math.random() * 1000);
	return `${adjective}_${noun}${number}`;
}

export function generateDatabaseName(): string {
	const prefixes = ['app', 'prod', 'dev', 'main', 'data', 'core'];
	const suffixes = ['db', 'data', 'store', 'cache', 'logs'];
	const prefix = prefixes[Math.floor(Math.random() * prefixes.length)];
	const suffix = suffixes[Math.floor(Math.random() * suffixes.length)];
	const number = Math.floor(Math.random() * 100);
	return `${prefix}_${suffix}${number}`;
}
