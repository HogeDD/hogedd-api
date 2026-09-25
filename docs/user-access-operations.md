# User権限の運用

管理GUIとrole変更APIを導入するまで、Productionのrole・status変更はDB管理者が明示的に実施します。公開APIやプロフィール入力からroleを変更してはいけません。

## 事前確認

1. 変更理由と承認者をIssueへ記録する。
2. Auth0 Dashboardで対象Userの`issuer`と`subject`を確認する。
3. emailだけで対象を決めない。emailは変更可能で、identityの識別子ではない。
4. NeonのSQL Editorまたは秘密情報を出力しない`psql` sessionを使用する。

## Roleの付与・剥奪

transaction内で対象と現在値を確認し、1件だけ更新します。

```sql
BEGIN;

SELECT id, auth_issuer, auth_subject, email, role, status
FROM users
WHERE auth_issuer = '<issuer>' AND auth_subject = '<subject>'
FOR UPDATE;

UPDATE users
SET role = '<owner|admin|member>', updated_at = NOW()
WHERE auth_issuer = '<issuer>' AND auth_subject = '<subject>';

COMMIT;
```

`SELECT`が0件または複数件の場合は`ROLLBACK`し、identityを再確認します。変更後は本人の既存sessionで`GET /v1/management/me`を確認し、期待する200または秘匿404になることを確認します。

## 緊急停止

不正利用の疑いがある場合は、role変更より先にHogeDD Userを停止します。

```sql
UPDATE users
SET status = 'disabled', updated_at = NOW()
WHERE auth_issuer = '<issuer>' AND auth_subject = '<subject>';
```

DBの`disabled`は既存Access Tokenにも即時適用されます。credential漏洩が疑われる場合は、あわせてAuth0 DashboardでUserをblockし、必要なsession・credentialを失効させます。復旧時は原因と本人確認を記録してから`active`へ戻します。

## 完了記録

- 対象の内部User ID
- 変更前後のrole・status
- 実施者と承認者
- 実施日時
- `/v1/management/me`の確認結果とrequest ID

email、Access Token、DB接続文字列などの秘密情報はIssueへ記録しません。
