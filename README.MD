# Protocast

Protocast brings Podcasts to the Open Social ATProtocol.

## Monorepo

This repository is a mono repo consisting of:
   * **Admin Frontend** - The frontend for creators and show managers to be able to add new feeds to the ATProto network, moderate existing shows/episodes, publish text-only posts for keeping the community in the loop, and more.
   * **Backend** - Shared go backend that will handle interacting with the ATProto and managing backend requests for our two frontend applications. Currently the only in progress, in the very early stages.
   * **Deploy** - Infrastucture for deploying the solution.
   * **Protocast Frontend** - Podcatcher app with traditional RSS feed functionality as well as enhanced functionality for Protocast-enabled shows.
   * **Shared Frontend** - For common UI

## Planned Features

The Admin frontend allows content creators to import their existing rss feeds into the Protocast network, or publish new content to a brand new podcast feed. It also allows creators to manage their posted content, as well as post updates directly to their community and manage moderation of the social functionality enabled by Protocast (primarily through removing comments that may appear on shows/individual episodes, though due to the nature of ATProtocol this would be gated by the webview the user is accessing the network through).

The Protocast frontend serves as a podcatcher for the network, while still supporting traditional RSS feed podcasts, albeit without the enhanced social functionality offered by Protocast. For Protocast-enabled podcasts users will be able to comment on shows, episodes and creator text-only posts. Additional functionality such as being able to see shows friends in your ATProtocol network listen to and track their progress is also planned, but the exact privacy model is to be determined.

