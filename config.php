<?php
/**
 * c00d IDE Configuration
 *
 * Edit your settings below. All options are documented inline.
 *
 * Tip: You can optionally create a config.local.php to override settings
 * without modifying this file (useful if you update c00d later).
 */

// Version constant
define('C00D_VERSION', trim(@file_get_contents(__DIR__ . '/VERSION') ?: '1.0.0'));

// =============================================================================
// PASSWORD
// =============================================================================
// Set your password here. If left empty, an auto-generated password is used
// (stored in data/.generated_password).

$_c00d_password = '';

if (empty($_c00d_password)) {
    $_c00d_password_file = __DIR__ . '/data/.generated_password';
    if (file_exists($_c00d_password_file)) {
        $_c00d_password = trim(@file_get_contents($_c00d_password_file) ?: '');
    } else {
        $_c00d_password = bin2hex(random_bytes(8));
        @mkdir(__DIR__ . '/data', 0755, true);
        @file_put_contents($_c00d_password_file, $_c00d_password);
        @chmod($_c00d_password_file, 0600);
    }
    if (empty($_c00d_password)) {
        $_c00d_password = substr(hash('sha256', __DIR__ . '|c00d-fallback-salt'), 0, 16);
    }
}

return [

    'password' => $_c00d_password,

    // Session lifetime in seconds (default: 24 hours)
    'session_lifetime' => 86400,

    // =========================================================================
    // FILE BROWSER
    // =========================================================================

    // Root folder for file browsing. Users cannot navigate above this path.
    'base_path' => dirname(__FILE__),

    'files' => [
        'show_hidden' => true,

        // Blocked folder names (security)
        'denied_paths' => ['.git', 'node_modules', 'vendor', '.env', '.env.local'],
    ],

    'security' => [
        // Restrict to specific IPs (empty = allow all)
        'allowed_ips' => [],

        // Require HTTPS
        'require_https' => false,
    ],

    // =========================================================================
    // AI
    // =========================================================================

    'ai' => [
        // Provider: 'c00d', 'claude-cli', 'anthropic', 'openai', 'ollama'
        'provider' => 'c00d',

        // c00d.com Pro license key — https://c00d.com/pro
        'license_key' => '',

        // API key (required for 'anthropic' or 'openai' providers)
        'api_key' => '',

        // Model name
        'model' => 'claude-sonnet-4-20250514',

        // Ollama server URL (only for 'ollama' provider)
        'ollama_url' => 'http://localhost:11434',

        // Claude CLI path (only for 'claude-cli' provider)
        'claude_cli_path' => 'claude',
    ],

    // =========================================================================
    // EDITOR
    // =========================================================================

    'editor' => [
        'theme' => 'vs-dark',        // vs-dark, vs-light, hc-black
        'font_size' => 14,
        'tab_size' => 4,
        'word_wrap' => 'on',         // on, off, wordWrapColumn, bounded
        'minimap' => true,
        'line_numbers' => 'on',      // on, off, relative
    ],

    // =========================================================================
    // TERMINAL
    // =========================================================================

    'terminal' => [
        'font_size' => 14,
        'history_size' => 1000,

        // WebSocket PTY terminal — gives a real shell but users can navigate
        // outside base_path. Only enable if you trust your users.
        'websocket_enabled' => false,
        'server_port' => 3456,
    ],
];
