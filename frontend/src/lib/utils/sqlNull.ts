/**
 * Utility functions for handling SQL null types from the backend
 * The backend uses sql.Null* types which serialize to JSON as objects with Valid and value fields
 */

interface SqlNullInt64 {
	Valid: boolean;
	Int64: number;
}

interface SqlNullString {
	Valid: boolean;
	String: string;
}

interface SqlNullBool {
	Valid: boolean;
	Bool: boolean;
}

interface SqlNullTime {
	Valid: boolean;
	Time: string;
}

/**
 * Extract a value from a sql.NullInt64 or return the direct value/default
 */
export function extractInt64(value: any, defaultValue: number): number;
export function extractInt64(value: any, defaultValue: string): string | number;
export function extractInt64(value: any, defaultValue: number | string = 'N/A'): number | string {
	if (value === null || value === undefined) {
		return defaultValue;
	}
	if (typeof value === 'object' && 'Valid' in value && 'Int64' in value) {
		return value.Valid ? value.Int64 : defaultValue;
	}
	return typeof value === 'number' ? value : defaultValue;
}

/**
 * Extract a value from a sql.NullString or return the direct value/default
 */
export function extractString(value: any, defaultValue: string = '-'): string {
	if (value === null || value === undefined) {
		return defaultValue;
	}
	if (typeof value === 'object' && 'Valid' in value && 'String' in value) {
		return value.Valid ? value.String : defaultValue;
	}
	return typeof value === 'string' ? value : defaultValue;
}

/**
 * Extract a value from a sql.NullBool or return the direct value/default
 */
export function extractBool(value: any, defaultValue: boolean = false): boolean {
	if (value === null || value === undefined) {
		return defaultValue;
	}
	if (typeof value === 'object' && 'Valid' in value && 'Bool' in value) {
		return value.Valid ? value.Bool : defaultValue;
	}
	return typeof value === 'boolean' ? value : defaultValue;
}

/**
 * Extract a value from a sql.NullTime or return the direct value/default
 */
export function extractTime(value: any, defaultValue: string | null = null): string | null {
	if (value === null || value === undefined) {
		return defaultValue;
	}
	if (typeof value === 'object' && 'Valid' in value && 'Time' in value) {
		return value.Valid ? value.Time : defaultValue;
	}
	return typeof value === 'string' ? value : defaultValue;
}

/**
 * Check if a sql.Null* value is valid (not null)
 */
export function isValidSqlNull(value: any): boolean {
	if (value === null || value === undefined) {
		return false;
	}
	if (typeof value === 'object' && 'Valid' in value) {
		return value.Valid;
	}
	return true;
}
