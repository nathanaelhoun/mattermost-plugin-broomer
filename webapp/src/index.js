import manifest from './manifest';

export default class Plugin {
    // eslint-disable-next-line no-unused-vars
    initialize(registry, store) {
        // @see https://developers.mattermost.com/extend/plugins/webapp/reference/
    }
}

window.registerPlugin(manifest.id, new Plugin());
