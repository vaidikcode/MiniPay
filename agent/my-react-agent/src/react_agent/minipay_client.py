"""Python HTTP client for MiniPay Go backend API.

Provides async methods to interact with payment operations, webhooks, and metrics.
"""

from __future__ import annotations

import httpx
from typing import Any, Optional
from dataclasses import dataclass


@dataclass
class ChargeResponse:
    """Response from a charge creation."""

    id: str
    amount: int
    currency: str
    customer: str
    status: str
    idempotency_key: str
    created_at: str


@dataclass
class RefundResponse:
    """Response from a refund operation."""

    id: str
    amount: int
    status: str
    refunded_at: str


@dataclass
class BalanceResponse:
    """Response from balance query."""

    successful_transactions: int
    refunded_transactions: int
    balance: int


@dataclass
class MetricsResponse:
    """Response from metrics endpoint."""

    total_charges: int
    total_refunds: int
    pending_webhooks: int
    delivered_webhooks: int
    failed_webhooks: int
    webhook_retries: int


class MiniPayClient:
    """Async HTTP client for MiniPay payment API."""

    def __init__(self, base_url: str = "http://localhost:8080", timeout: float = 30.0):
        """Initialize the MiniPay client.

        Args:
            base_url: Base URL of the MiniPay backend (default: http://localhost:8080)
            timeout: Request timeout in seconds (default: 30.0)
        """
        self.base_url = base_url.rstrip("/")
        self.timeout = timeout
        self.client = httpx.AsyncClient(timeout=timeout)

    async def create_charge(
        self,
        amount: int,
        currency: str = "usd",
        customer: str = "",
        idempotency_key: Optional[str] = None,
    ) -> ChargeResponse:
        """Create a new charge.

        Args:
            amount: Amount in cents (e.g., 1000 = $10.00)
            currency: Currency code (default: usd)
            customer: Customer identifier
            idempotency_key: Optional idempotency key for deduplication

        Returns:
            ChargeResponse with transaction details

        Raises:
            httpx.HTTPError: If the request fails
        """
        url = f"{self.base_url}/api/v1/charges"
        payload = {
            "amount": amount,
            "currency": currency,
            "customer": customer,
        }
        headers = {}
        if idempotency_key:
            headers["Idempotency-Key"] = idempotency_key

        response = await self.client.post(url, json=payload, headers=headers)
        response.raise_for_status()
        data = response.json()
        return ChargeResponse(**data)

    async def refund_transaction(self, transaction_id: str) -> RefundResponse:
        """Refund an existing transaction.

        Args:
            transaction_id: The transaction ID to refund

        Returns:
            RefundResponse with refund details

        Raises:
            httpx.HTTPError: If the request fails
        """
        url = f"{self.base_url}/api/v1/refunds"
        payload = {"transaction_id": transaction_id}
        response = await self.client.post(url, json=payload)
        response.raise_for_status()
        data = response.json()
        return RefundResponse(**data)

    async def get_balance(self) -> BalanceResponse:
        """Get current account balance.

        Returns:
            BalanceResponse with balance details

        Raises:
            httpx.HTTPError: If the request fails
        """
        url = f"{self.base_url}/api/v1/balance"
        response = await self.client.get(url)
        response.raise_for_status()
        data = response.json()
        return BalanceResponse(**data)

    async def get_metrics(self) -> MetricsResponse:
        """Get system metrics.

        Returns:
            MetricsResponse with system metrics

        Raises:
            httpx.HTTPError: If the request fails
        """
        url = f"{self.base_url}/metrics"
        response = await self.client.get(url)
        response.raise_for_status()
        data = response.json()
        return MetricsResponse(**data)

    async def health_check(self) -> bool:
        """Check if the backend is healthy.

        Returns:
            True if backend is healthy, False otherwise
        """
        try:
            url = f"{self.base_url}/health"
            response = await self.client.get(url, timeout=5.0)
            return response.status_code == 200
        except (httpx.RequestError, httpx.TimeoutException):
            return False

    async def close(self) -> None:
        """Close the HTTP client."""
        await self.client.aclose()

    async def __aenter__(self) -> MiniPayClient:
        """Context manager entry."""
        return self

    async def __aexit__(self, exc_type: Any, exc_val: Any, exc_tb: Any) -> None:
        """Context manager exit."""
        await self.close()
