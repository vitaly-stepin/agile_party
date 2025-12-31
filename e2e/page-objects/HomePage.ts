import { Page, Locator } from '@playwright/test';

export class HomePage {
  readonly page: Page;
  readonly createRoomTab: Locator;
  readonly joinRoomTab: Locator;
  readonly roomNameInput: Locator;
  readonly nicknameInputCreate: Locator;
  readonly nicknameInputJoin: Locator;
  readonly roomIdInput: Locator;
  readonly createRoomButton: Locator;
  readonly joinRoomButton: Locator;

  constructor(page: Page) {
    this.page = page;
    this.createRoomTab = page.getByText('Create Room').first();
    this.joinRoomTab = page.getByText('Join Room').first();
    this.roomNameInput = page.locator('#room-name');
    this.nicknameInputCreate = page.locator('#nickname').first();
    this.nicknameInputJoin = page.locator('#join-nickname');
    this.roomIdInput = page.locator('#room-id');
    this.createRoomButton = page.locator('form').getByRole('button', { name: /create room/i });
    this.joinRoomButton = page.locator('form').getByRole('button', { name: /join room/i });
  }

  async goto() {
    await this.page.goto('/');
  }

  async createRoom(roomName: string, nickname: string) {
    await this.createRoomTab.click();
    await this.roomNameInput.fill(roomName);
    await this.nicknameInputCreate.fill(nickname);
    await this.createRoomButton.click();
  }

  async joinRoom(roomId: string, nickname: string) {
    await this.joinRoomTab.click();
    await this.roomIdInput.fill(roomId);
    await this.nicknameInputJoin.fill(nickname);
    await this.joinRoomButton.click();
  }
}
