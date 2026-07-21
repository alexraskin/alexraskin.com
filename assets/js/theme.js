// Applied before first paint (script is not deferred) to avoid a flash of the
// wrong theme. With no stored preference we leave data-theme unset so the CSS
// prefers-color-scheme rules decide.
(() => {
    const COOKIE = 'theme';

    const getCookie = (name) => {
        const parts = `; ${document.cookie}`.split(`; ${name}=`);
        return parts.length === 2 ? parts.pop().split(';').shift() : null;
    };

    const setCookie = (name, value, days) => {
        const date = new Date();
        date.setTime(date.getTime() + days * 24 * 60 * 60 * 1000);
        document.cookie = `${name}=${value};expires=${date.toUTCString()};path=/;samesite=lax`;
    };

    const systemTheme = () =>
        window.matchMedia('(prefers-color-scheme: dark)').matches ? 'dark' : 'light';

    const saved = getCookie(COOKIE);
    if (saved === 'dark' || saved === 'light') {
        document.documentElement.setAttribute('data-theme', saved);
    }

    document.addEventListener('DOMContentLoaded', () => {
        // Sidebar (desktop) and header (mobile) each render a toggle.
        document.querySelectorAll('.js-theme-toggle').forEach((toggle) => {
            toggle.addEventListener('click', (e) => {
                e.preventDefault();
                const current = document.documentElement.getAttribute('data-theme') || systemTheme();
                const next = current === 'dark' ? 'light' : 'dark';
                document.documentElement.setAttribute('data-theme', next);
                setCookie(COOKIE, next, 365);
            });
        });
    });
})();
