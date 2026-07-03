(function () {
    const mountPath = '/mmh';

    function isMounted() {
        return window.location.pathname === mountPath || window.location.pathname.startsWith(`${mountPath}/`);
    }

    function appPath(path) {
        if (!path) return isMounted() ? `${mountPath}/` : '/';
        if (/^(https?:|data:|blob:|#)/.test(path)) return path;

        const normalized = path.startsWith('/') ? path : `/${path}`;
        if (normalized === mountPath || normalized.startsWith(`${mountPath}/`)) return normalized;

        return isMounted() ? `${mountPath}${normalized}` : normalized;
    }

    window.appPath = appPath;
})();
