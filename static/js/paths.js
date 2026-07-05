(function () {
    const mountPath = '/mmh';
    const wordpressPageId = '310904';
    const routeKey = 'mmh:lastRoute';

    function isMounted() {
        return window.location.pathname === mountPath || window.location.pathname.startsWith(`${mountPath}/`);
    }

    function isWordPressShell() {
        return new URLSearchParams(window.location.search).get('page_id') === wordpressPageId;
    }

    function normalizeRoute(pathname) {
        let path = pathname || '/';
        if (path === mountPath) return '/';
        if (path.startsWith(`${mountPath}/`)) path = path.slice(mountPath.length);
        if (!path.startsWith('/')) path = `/${path}`;
        if (path === '') path = '/';
        if (path === '/' || /^\/(images|photos|my-media|profile\/edit|upload|admin|login|register)$/.test(path) || /^\/photo\/\d+$/.test(path)) {
            return path;
        }
        return null;
    }

    function rememberRoute(route) {
        if (!route) return;
        try {
            localStorage.setItem(routeKey, route);
        } catch (error) {
            // Storage can be disabled in some embedded browsers.
        }
    }

    function rememberedRoute() {
        try {
            return normalizeRoute(localStorage.getItem(routeKey) || '') || '';
        } catch (error) {
            return '';
        }
    }

    function appPath(path) {
        if (!path) return isMounted() ? `${mountPath}/` : '/';
        if (/^(https?:|data:|blob:|#)/.test(path)) return path;

        const normalized = path.startsWith('/') ? path : `/${path}`;
        if (normalized === mountPath || normalized.startsWith(`${mountPath}/`)) return normalized;

        if (isMounted() || isWordPressShell()) return `${mountPath}${normalized}`;
        return normalized;
	}

    function appRouteUrl(route) {
        const normalized = normalizeRoute(route) || '/';
        if (!isWordPressShell()) return appPath(normalized);

        const url = new URL(window.location.href);
        url.hash = '';
        return url.pathname + url.search;
    }

    async function fetchRouteHTML(route) {
        const candidates = [];
        const primary = appPath(route);
        candidates.push(primary);
        if (!primary.startsWith(`${mountPath}/`)) candidates.push(`${mountPath}${route}`);
        candidates.push(route);

        const uniqueCandidates = [...new Set(candidates)];
        for (const candidate of uniqueCandidates) {
            try {
                const response = await fetch(candidate, { credentials: 'include' });
                if (!response.ok) continue;
                return await response.text();
            } catch (error) {
                // Try the next route candidate.
            }
        }
        return '';
    }

    async function renderWordPressRoute(route, pushState = true) {
        const normalized = normalizeRoute(route);
        if (!normalized) return false;

        rememberRoute(normalized);
        if (pushState) {
            window.history.pushState({ mmhRoute: normalized }, '', appRouteUrl(normalized));
        }

        const html = await fetchRouteHTML(normalized);
        if (!html) {
            window.location.href = appPath(normalized);
            return true;
        }

        const markedHTML = html.replace(/<html([^>]*)>/i, `<html$1 data-mmh-restored-route="${normalized.replace(/"/g, '&quot;')}">`);
        document.open();
        document.write(markedHTML);
        document.close();
        return true;
    }

    function appNavigate(route) {
        const normalized = normalizeRoute(route);
        if (isWordPressShell() && normalized) {
            renderWordPressRoute(normalized);
            return;
        }
        window.location.href = appPath(route);
    }

    function bindWordPressRouteLinks() {
        if (!isWordPressShell()) return;

        document.addEventListener('click', (event) => {
            const link = event.target.closest('a[href]');
            if (!link || link.target || event.metaKey || event.ctrlKey || event.shiftKey || event.altKey) return;

            const url = new URL(link.getAttribute('href'), window.location.href);
            if (url.origin !== window.location.origin) return;
            const route = normalizeRoute(url.pathname);
            if (!route) return;

            event.preventDefault();
            renderWordPressRoute(route);
        });
    }

    function restoreWordPressRoute() {
        if (!isWordPressShell()) return;

        const target = rememberedRoute();
        if (!target || target === '/') return;
        if (document.documentElement.dataset.mmhRestoredRoute === target) return;

        renderWordPressRoute(target, false);
    }

    function bindWordPressHistory() {
        if (!isWordPressShell()) return;

        window.addEventListener('popstate', () => {
            const target = normalizeRoute(history.state?.mmhRoute || '') || rememberedRoute() || '/';
            if (target === '/') {
                rememberRoute('/');
                return;
            }
            renderWordPressRoute(target, false);
        });
    }

    const currentRoute = normalizeRoute(window.location.pathname);
    if (currentRoute && isMounted()) {
        rememberRoute(currentRoute);
    }

    window.appPath = appPath;
    window.appRouteUrl = appRouteUrl;
    window.appNavigate = appNavigate;
    bindWordPressRouteLinks();
    bindWordPressHistory();
    restoreWordPressRoute();
})();
