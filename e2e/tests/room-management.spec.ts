import { test, expect } from '../fixtures/database.fixture';
import { HomePage } from '../page-objects/HomePage';
import { RoomPage } from '../page-objects/RoomPage';

test.describe('Room Management', () => {
  test('should leave room and navigate to home', async ({ page }) => {
    const homePage = new HomePage(page);
    const roomPage = new RoomPage(page);

    // Setup: Create room
    await homePage.goto();
    const roomName = `E2E Test Room ${Date.now()}`;
    const nickname = 'Test User';
    await homePage.createRoom(roomName, nickname);

    // Verify in room
    await expect(page).toHaveURL(/\/room\/[a-f0-9-]+/);
    await roomPage.waitForWebSocketConnection();

    // Action: Leave room
    await roomPage.leaveRoom();

    // Assert: Navigated to home page
    await expect(page).toHaveURL('/');
    await expect(homePage.createRoomTab).toBeVisible();
  });

  test('should remove user from other participants when leaving', async ({ page, context }) => {
    const homePage = new HomePage(page);
    const roomPage = new RoomPage(page);

    // Setup: Create room with first user
    await homePage.goto();
    const roomName = `E2E Test Room ${Date.now()}`;
    const firstUserNickname = 'First User';
    await homePage.createRoom(roomName, firstUserNickname);

    await roomPage.waitForWebSocketConnection();
    await roomPage.waitForUserToAppear(firstUserNickname);

    const roomId = await roomPage.getRoomId();

    // Setup: Second user joins
    const secondUserPage = await context.newPage();
    const secondUserHomePage = new HomePage(secondUserPage);
    const secondUserRoomPage = new RoomPage(secondUserPage);

    await secondUserHomePage.goto();
    const secondUserNickname = 'Second User';
    await secondUserHomePage.joinRoom(roomId, secondUserNickname);

    await secondUserRoomPage.waitForWebSocketConnection();

    // Verify both users see each other
    await roomPage.waitForUserToAppear(secondUserNickname, 10000);
    await secondUserRoomPage.waitForUserToAppear(firstUserNickname, 10000);

    // Action: Second user leaves room
    await secondUserRoomPage.leaveRoom();

    // Assert: First user no longer sees second user
    await roomPage.waitForUserToDisappear(secondUserNickname, 10000);

    // First user should still see themselves
    await expect(roomPage.getUserInList(firstUserNickname)).resolves.toBeVisible();

    // Cleanup
    await secondUserPage.close();
  });

  test('should rejoin room after leaving', async ({ page }) => {
    const homePage = new HomePage(page);
    const roomPage = new RoomPage(page);

    // Setup: Create room
    await homePage.goto();
    const roomName = `E2E Test Room ${Date.now()}`;
    const nickname = 'Test User';
    await homePage.createRoom(roomName, nickname);

    await roomPage.waitForWebSocketConnection();
    await roomPage.waitForUserToAppear(nickname);

    // Get room ID before leaving
    const roomId = await roomPage.getRoomId();

    // Action: Leave room
    await roomPage.leaveRoom();
    await expect(page).toHaveURL('/');

    // Action: Rejoin same room
    await homePage.joinRoom(roomId, nickname);

    // Assert: User is back in room
    await expect(page).toHaveURL(/\/room\/[a-f0-9-]+/);
    await expect(roomPage.roomTitle).toContainText(roomName);
    await roomPage.waitForWebSocketConnection();
    await roomPage.waitForUserToAppear(nickname, 10000);
  });
});
